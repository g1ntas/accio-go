package main

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"os"
	"strings"

	"github.com/g1ntas/accio/generator"
	"github.com/g1ntas/accio/generator/blueprint"
	"github.com/g1ntas/accio/internal/fs"
	"github.com/g1ntas/accio/internal/logger"
	"github.com/g1ntas/accio/internal/manifest"
)

const manifestFilename = ".accio.toml"

var runCmd = &cobra.Command{
	Use:   "run [generator]",
	Short: "Run a generator from directory or git repository",
	Long: `Executes generator writing all generated files at the current 
directory, unless specified otherwise. If the generator has 
any prompts configured, then they will be prompted first.

Command accepts a single required argument specifying the 
location of the generator. Location can be either a path 
to a local directory or an URL to a git repository.

Git repository URLs:
  HTTP, HTTPS, SSH, GIT, and SCP-style Git URLs are 
  supported. In case, git provider is a known one (github.com,
  bitbucket.org, gitlab.com, gitea.com), URL can be specified 
  without scheme and it will be resolved automatically.

  For authentication with credentials, specify username and 
  password separated by a colon in URL like in the example 
  below. If git provider supports access tokens, for security
  reasons, it's recommended to use one with repository read 
  access, instead of a real password.
  Example: username:password@github.com/owner/repository.

  For authentication with an SSH access key, SSH format URL 
  should be used. Access keys are detected by the active 
  ssh-agent process. If the ssh-agent is not running, an error
  will be returned. Also, the ssh key must be registered with 
  an active ssh-agent for it to be detected, if that's not the 
  case yet, the key can be registered with the following 
  command: 'ssh-add /path/to/private/key'.
  Example: git@github.com:owner/repository

  Subdirectories are supported and can be specified after 
  double-slash '//'. In the case of know git provider, only a 
  single slash '/' can be used.
  Examples:
  https://host.com/repository//subdirectory
  github.com/g1ntas/accio/examples/open-source-license

  Git references can be specified at the end of the URL as a 
  hash fragment and have to be valid full git reference. By 
  default, HEAD reference is used.
  Examples:
  github.com/owner/repo#refs/tags/1.0.0
  github.com/owner/repo#refs/branch/some-branch
  github.com/owner/repo#HEAD
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		treeReader, gen, err := fetchGenerator(args[0])
		if err != nil {
			return err
		}
		writeDir, err := workingDir(cmd)
		if err != nil {
			return err
		}
		env.log.Debug("working directory: ", writeDir)
		data, err := gen.PromptAll(env.prompter)
		if err != nil {
			return err
		}
		parser, err := blueprint.NewParser(data, logger.NewFromLogger(env.log, "blueprint"))
		if err != nil {
			return err
		}
		options := append([]generator.OptionFn{
			generator.OnFileExists(existingFileHandler(cmd)),
			generator.WithLogger(logger.NewFromLogger(env.log, "generator")),
			generator.IgnorePath(".git"),
			generator.IgnorePath(manifestFilename),
		}, ignoredPaths(gen)...)
		if getBoolFlag(cmd, "ignore-errors") {
			options = append(options, generator.SkipErrors)
		}
		runner := generator.NewRunner(filesystem(cmd), parser, writeDir, options...)
		env.log.Info("Running...")
		err = runner.Run(treeReader)
		if err != nil {
			return err
		}
		env.log.Info("Done.")
		return nil
	},
}

func init() {
	runCmd.SetHelpFunc(generatorHelpFunc)
	runCmd.Flags().Bool("dry", false, "Run without writing to filesystem")
	runCmd.Flags().BoolP("force", "f", false, "Overwrite existing paths without asking confirmation")
	runCmd.Flags().BoolP("help", "h", false, "Show help")
	runCmd.Flags().BoolP("ignore-errors", "i", false, "Ignore errors for files being generated")
	runCmd.Flags().StringP("working-dir", "w", "", "Specify working directory")
	rootCmd.AddCommand(runCmd)
}

func fetchGenerator(src string) (generator.FileTreeReader, *manifest.Generator, error) {
	treeReader, err := urlToTreeReader(src)
	if err != nil {
		return nil, nil, err
	}
	gen, err := readManifest(treeReader)
	if err != nil {
		return nil, nil, err
	}
	env.log.Debug("parsed manifest: ", gen)
	return treeReader, gen, nil
}

func readManifest(r generator.FileTreeReader) (*manifest.Generator, error) {
	b, err := r.ReadFile(manifestFilename)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}
	gen, err := manifest.ReadToml(b)
	if err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	return gen, nil
}

func urlToTreeReader(src string) (generator.FileTreeReader, error) {
	info, err := env.fs.Stat(src)
	if err == nil && info.IsDir() {
		env.log.Debug("reading generator from local directory")
		return fs.NewAferoFileTreeReader(env.fs, src), nil
	}
	env.log.Debug("reading generator from remote git repository")
	return env.git.Get(src)
}

func ignoredPaths(gen *manifest.Generator) (paths []generator.OptionFn) {
	for _, p := range gen.Ignore {
		paths = append(paths, generator.IgnorePath(p))
	}
	return paths
}

// generatorHelpFunc defines run command's help behaviour.
// When --help flag is provided together with at least single argument,
// generator will be parsed and help text from configuration file will be shown.
// In case of no arguments, default help behaviour will be executed.
func generatorHelpFunc(cmd *cobra.Command, args []string) {
	if err := cmd.ValidateArgs(cmd.Flags().Args()); err != nil {
		cmd.Root().HelpFunc()(cmd, args)
		return
	}
	_, gen, err := fetchGenerator(args[1])
	if err != nil {
		printErr(err)
		fmt.Println(cmd.UsageString())
		os.Exit(1)
	}
	helpCmd := &cobra.Command{
		Use:   args[1],
		Short: "",
		Long:  buildGeneratorHelp(gen),
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	cmd.AddCommand(helpCmd)
	helpCmd.Flags().AddFlagSet(cmd.Flags())
	helpCmd.SetHelpFunc(cmd.Root().HelpFunc())
	helpCmd.HelpFunc()(helpCmd, args)
}

func filesystem(cmd *cobra.Command) afero.Afero {
	if getBoolFlag(cmd, "dry") {
		env.log.Debug("running in dry mode")
		roBase := afero.NewReadOnlyFs(env.fs.Fs)
		ufs := afero.NewCopyOnWriteFs(roBase, afero.NewMemMapFs())
		return afero.Afero{Fs: ufs}
	}
	return env.fs
}

func existingFileHandler(cmd *cobra.Command) generator.OnExistsFn {
	force := getBoolFlag(cmd, "force")
	return func(path string) bool {
		if force {
			env.log.Info("Force overwriting file at ", path)
			return true
		}
		fmt.Printf("File at path %q already exists\n", path)
		overwrite, err := env.prompter.Confirm("Do you want to overwrite it?", "")
		if err != nil {
			env.log.Info("ERROR: ", err)
			return false
		}
		return overwrite
	}
}

func workingDir(cmd *cobra.Command) (string, error) {
	dir := getStringFlag(cmd, "working-dir")
	if dir != "" {
		return dir, nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return dir, nil
}

func buildGeneratorHelp(gen *manifest.Generator) string {
	help := strings.TrimSpace(gen.Help)
	if len(gen.Prompts) == 0 {
		return help
	}
	help += "\n\nPrompts:\n"
	for name, pr := range gen.Prompts {
		help += fmt.Sprintf("[%s]\n", name)
		switch h := pr.Help(); {
		case len(h) > 0:
			help += h
		default:
			help += "This prompt has no description"
		}
		help += "\n\n"
	}
	return strings.TrimSpace(help)
}
