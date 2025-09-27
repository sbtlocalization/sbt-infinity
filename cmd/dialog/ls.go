package dialog

import (
	"fmt"
	"path/filepath"

	"github.com/sbtlocalization/infinity-tools/dialog"
	"github.com/sbtlocalization/infinity-tools/fs"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func NewLsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <path to chitin.key> [dialog files...]",
		Aliases: []string{"ls"},
		Short:   "List dialogs from the game",
		Long: `List all dialogs or specific dialog files from the game.
		Reads the game structure from chitin.key file and dialog.tlk file, and optionally lists
		only specified dialog files (e.g., ABISHAB.DLG, DMORTE.DLG).`,
		Args: cobra.MinimumNArgs(1),
		RunE: runLs,
	}

	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().StringP("tlk", "t", "", "Path to dialog.tlk file")

	return cmd
}

func runLs(cmd *cobra.Command, args []string) error {
	keyPath := args[0]
	dialogFiles := args[1:]

	jsonOutput, _ := cmd.Flags().GetBool("json")
	tlkPath, _ := cmd.Flags().GetString("tlk")

	osFs := afero.NewOsFs()

	var tlkFs afero.Fs
	if tlkPath == "" {
		tlkFs = afero.NewBasePathFs(osFs, filepath.Join(filepath.Dir(keyPath)))
		tlkPath = "lang/en_US/dialog.tlk"
	} else {
		tlkFs = osFs
	}

	dlgFs := fs.NewInfinityFs(keyPath, fs.FileType_DLG)

	dc := dialog.NewDialogBuilder(dlgFs, tlkFs)

	if len(dialogFiles) == 0 {
		dir, err := dlgFs.Open("DLG")
		if err != nil {
			return fmt.Errorf("unable to list existing DLG files: %v", err)
		}
		dialogFiles, err = dir.Readdirnames(0)
		if err != nil {
			return fmt.Errorf("unable to read dialog directory names: %v", err)
		}
	}

	for _, df := range dialogFiles {
		dlg, err := dc.LoadAllRootStates(df)
		if err != nil {
			return fmt.Errorf("error loading dialogs: %v", err)
		}

		_ = jsonOutput

		for _, d := range dlg.Dialogs {
			fmt.Println(d.Id)
		}
	}

	return nil
}
