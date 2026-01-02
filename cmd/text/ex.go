// SPDX-FileCopyrightText: © 2025 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

package text

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
	"github.com/samber/lo"
	"github.com/sbtlocalization/sbt-infinity/config"
	"github.com/sbtlocalization/sbt-infinity/dialog"
	"github.com/sbtlocalization/sbt-infinity/fs"
	p "github.com/sbtlocalization/sbt-infinity/parser"
	"github.com/sbtlocalization/sbt-infinity/text"
	"github.com/sbtlocalization/sbt-infinity/utils"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func NewExCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "export [ID...]",
		Aliases: []string{"ex"},
		Short:   "Export textual resources from the game as xlsx",
		Long: `Export all textual resources or specific IDs from the game.
Reads the texts from dialog.tlk file, and optionally extracts only specified
text IDs (e.g., 1234, 5678).`,
		Args: cobra.MinimumNArgs(0),
		RunE: runEx,
	}

	cmd.Flags().StringP("output", "o", "dialog.xlsx", "output xlsx file `path`")
	cmd.Flags().BoolP("verbose", "v", false, "enable verbose output")
	cmd.Flags().String("dlg-base-url", "", "base `URL` for dialog references (overrides config)")
	cmd.Flags().StringSlice("context-from", []string{}, "load context from types of files. Use 'all' to include all types.\nUse 'bif `types`' command to see all types.")
	cmd.Flags().String("timestamps-from", "", "CSV file `path` containing timestamps to include in the export")

	return cmd
}

func runEx(cmd *cobra.Command, args []string) error {
	tlkPath, _ := cmd.Flags().GetString("tlk")
	lang, _ := cmd.Flags().GetString("lang")
	feminine, _ := cmd.Flags().GetBool("feminine")
	verbose, _ := cmd.Flags().GetBool("verbose")
	baseUrl, _ := config.ResolveDialogBaseUrl(cmd)
	contextFrom, _ := cmd.Flags().GetStringSlice("context-from")
	timestampsFrom, _ := cmd.Flags().GetString("timestamps-from")

	outputPath, _ := cmd.Flags().GetString("output")
	if cmd.Flags().Changed("output") && !strings.HasSuffix(strings.ToLower(outputPath), ".xlsx") {
		outputPath = outputPath + ".xlsx"
	}

	keyPath, err := config.ResolveKeyPath(cmd)
	if err != nil {
		return err
	}

	osFs := afero.NewOsFs()

	var tlkFs afero.Fs
	if !cmd.Flags().Changed("tlk") {
		tlkFs = afero.NewBasePathFs(osFs, filepath.Dir(keyPath))
		if feminine {
			tlkPath = filepath.Join("lang", lang, "dialogf.tlk")
		} else {
			tlkPath = filepath.Join("lang", lang, "dialog.tlk")
		}
	} else {
		tlkFs = osFs
	}

	if verbose {
		fmt.Print("loading TLK file... ")
	}
	tlkFile, err := p.ReadTlkFile(tlkFs, tlkPath)
	if err != nil {
		return err
	}
	collection := text.NewTextCollection(tlkFile.Tlk)
	tlkFile.Close()
	if verbose {
		fmt.Println("done.")
	}

	contextTypes := []fs.FileType{
		fs.FileType_2DA,
		fs.FileType_ARE,
		fs.FileType_CHU,
		fs.FileType_CRE,
		fs.FileType_DLG,
		fs.FileType_EFF,
		fs.FileType_ITM,
		fs.FileType_PRO,
		fs.FileType_SPL,
		fs.FileType_STO,
		fs.FileType_WMP,
	}
	if !slices.Contains(contextFrom, "all") {
		contextTypes = lo.UniqMap(contextFrom, utils.Iteratee(fs.FileTypeFromExtension))
	}

	infFs := fs.NewInfinityFs(keyPath)

	for _, t := range contextTypes {
		switch t {
		case fs.FileType_2DA:
			err = process2daFiles(collection, infFs, verbose)
			if err != nil {
				fmt.Println("warning: unable to process 2DA files:", err)
			}
		case fs.FileType_ARE:
			err = processAreas(collection, infFs, verbose)
			if err != nil {
				fmt.Println("warning: unable to process areas:", err)
			}
		case fs.FileType_CHU:
			err = processUiScreens(collection, infFs, verbose)
			if err != nil {
				fmt.Println("warning: unable to process UI screens:", err)
			}
		case fs.FileType_CRE:
			err = processCreatures(collection, infFs, verbose)
			if err != nil {
				fmt.Println("warning: unable to process creatures:", err)
			}
		case fs.FileType_DLG:
			err = processDialogs(collection, infFs, baseUrl, verbose)
			if err != nil {
				fmt.Println("warning: unable to process dialogs:", err)
			}
		case fs.FileType_EFF:
			err = processEffects(collection, infFs, verbose)
			if err != nil {
				fmt.Println("warning: unable to process effects:", err)
			}
		case fs.FileType_ITM:
			err = processItems(collection, infFs, verbose)
			if err != nil {
				fmt.Println("warning: unable to process items:", err)
			}
		case fs.FileType_PRO:
			err = processProjectiles(collection, infFs, verbose)
			if err != nil {
				fmt.Println("warning: unable to process projectiles:", err)
			}
		case fs.FileType_SPL:
			err = processSpells(collection, infFs, verbose)
			if err != nil {
				fmt.Println("warning: unable to process spells:", err)
			}
		case fs.FileType_STO:
			err = processStores(collection, infFs, verbose)
			if err != nil {
				fmt.Println("warning: unable to process stores:", err)
			}
		case fs.FileType_WMP:
			err = processWorldMaps(collection, infFs, verbose)
			if err != nil {
				fmt.Println("warning: unable to process world maps:", err)
			}
		default:
			continue
		}
	}

	collection.FillKnownContext()

	// Load timestamps from CSV if provided
	var timestamps map[uint32]int64
	if timestampsFrom != "" {
		if verbose {
			fmt.Print("loading timestamps from CSV... ")
		}

		timestampEntries, err := loadTimestamps(timestampsFrom)
		if err != nil {
			return err
		}

		if verbose {
			fmt.Printf("done (%d entries).\n", len(timestampEntries))
		}

		// Validate and extract timestamps
		timestamps = make(map[uint32]int64)
		csvIds := make(map[uint32]struct{})

		for id, entry := range timestampEntries {
			csvIds[id] = struct{}{}
			timestamps[id] = entry.Timestamp

			// Check if ID exists in collection and text matches
			if tlkEntry, ok := collection.Entries[id]; ok {
				if tlkEntry.Text != entry.Text {
					fmt.Printf("warning: ID %d text mismatch - TLK: %q, CSV: %q\n", id, tlkEntry.Text, entry.Text)
				}
			} else {
				fmt.Printf("warning: CSV contains ID %d which is not in the TLK file\n", id)
			}
		}

		// Check for IDs in collection that are not in CSV
		for id := range collection.Entries {
			if _, ok := csvIds[id]; !ok {
				fmt.Printf("warning: ID %d not found in timestamps CSV\n", id)
			}
		}
	}

	err = collection.ExportToXlsx(outputPath, timestamps)
	if err != nil {
		return err
	}

	return nil
}

func processDialogs(collection *text.TextCollection, infFs afero.Fs, baseUrl string, verbose bool) error {
	dlgBuilder := dialog.NewDialogBuilder(infFs, nil, false, verbose)
	dir, err := infFs.Open("DLG")
	if err != nil {
		return fmt.Errorf("unable to list existing DLG files: %v", err)
	}
	defer dir.Close()
	dialogFiles, err := dir.Readdirnames(0)
	if err != nil {
		return fmt.Errorf("unable to read dialog directory names: %v", err)
	}

	total := len(dialogFiles)

	if verbose {
		fmt.Print("extracting context from dialogs...")
	}

	for _, df := range dialogFiles {
		dc, err := dlgBuilder.LoadAllDialogs("", df)
		if err != nil {
			return fmt.Errorf("error loading dialogs: %v", err)
		}

		collection.LoadContextFromDialogs(baseUrl, dc)
	}

	if verbose {
		fmt.Printf(" done (%d files).\n", total)
	}

	return nil
}

func processCreatures(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	dir, err := infFs.Open("CRE")
	if err != nil {
		return fmt.Errorf("unable to list existing CRE files: %v", err)
	}
	defer dir.Close()

	creFiles, err := dir.Readdirnames(0)
	if err != nil {
		return fmt.Errorf("unable to read CRE directory names: %v", err)
	}

	var ids *p.Ids

	sndslot, err := infFs.Open("SNDSLOT.IDS")
	if err != nil {
		fmt.Println("warning: unable to open SNDSLOT.IDS:", err)
	} else {
		ids, err = p.ParseIds(sndslot)
		sndslot.Close()
		if err != nil {
			fmt.Println("warning: unable to parse SNDSLOT.IDS:", err)
		}
	}

	total := len(creFiles)

	if verbose {
		fmt.Print("extracting context from creatures...")
	}

	for _, cf := range creFiles {
		creFile, err := infFs.Open(cf)
		if err != nil {
			return fmt.Errorf("unable to open CRE file %q: %v", cf, err)
		}
		defer creFile.Close()

		cre := p.NewCre()
		stream := kaitai.NewStream(creFile)
		err = cre.Read(stream, nil, cre)
		if err != nil {
			return fmt.Errorf("unable to parse CRE file %q: %v", cf, err)
		}

		collection.LoadContextFromCreature(cf, cre, ids)
	}

	if verbose {
		fmt.Printf(" done (%d files).\n", total)
	}

	return nil
}

func processFiles(
	infFs afero.Fs,
	verbose bool,
	dirName string,
	entityName string,
	processFile func(filename string, stream *kaitai.Stream) error,
) error {
	dir, err := infFs.Open(dirName)
	if err != nil {
		return fmt.Errorf("unable to list existing %s files: %v", dirName, err)
	}
	defer dir.Close()

	files, err := dir.Readdirnames(0)
	if err != nil {
		return fmt.Errorf("unable to read %s directory names: %v", dirName, err)
	}

	total := len(files)
	processed := 0
	hasWarnings := false

	if verbose {
		fmt.Printf("extracting context from %s...", entityName)
	}

	for _, f := range files {
		file, err := infFs.Open(f)
		if err != nil {
			if verbose {
				if !hasWarnings {
					fmt.Println()
					hasWarnings = true
				}
				fmt.Printf("  warning: unable to open %s file %q: %v. skipping...\n", dirName, f, err)
			}
			continue
		}

		stream := kaitai.NewStream(file)
		if err := processFile(f, stream); err != nil {
			if verbose {
				if !hasWarnings {
					fmt.Println()
					hasWarnings = true
				}
				fmt.Printf("  warning: unable to parse %s file %q: %v. skipping...\n", dirName, f, err)
			}
		} else {
			processed++
		}
		file.Close()
	}

	if verbose {
		if processed == total {
			fmt.Printf(" done (%d files).\n", total)
		} else {
			fmt.Printf("done (%d/%d files).\n", processed, total)
		}
	}

	return nil
}

func processUiScreens(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	return processFiles(infFs, verbose, "CHU", "UI screens", func(filename string, stream *kaitai.Stream) error {
		chu := p.NewChu()
		if err := chu.Read(stream, nil, chu); err != nil {
			return err
		}
		collection.LoadContextFromUiScreens(filename, chu)
		return nil
	})
}

func processWorldMaps(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	return processFiles(infFs, verbose, "WMP", "world maps", func(filename string, stream *kaitai.Stream) error {
		wmp := p.NewWmp()
		if err := wmp.Read(stream, nil, wmp); err != nil {
			return err
		}
		collection.LoadContextFromWorldMaps(filename, wmp)
		return nil
	})
}

func processAreas(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	return processFiles(infFs, verbose, "ARE", "areas", func(filename string, stream *kaitai.Stream) error {
		are := p.NewAre()
		if err := are.Read(stream, nil, are); err != nil {
			return err
		}
		collection.LoadContextFromArea(filename, are)
		return nil
	})
}

func processItems(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	return processFiles(infFs, verbose, "ITM", "items", func(filename string, stream *kaitai.Stream) error {
		itm := p.NewItm()
		if err := itm.Read(stream, nil, itm); err != nil {
			return err
		}
		collection.LoadContextFromItem(filename, itm)
		return nil
	})
}

func processProjectiles(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	return processFiles(infFs, verbose, "PRO", "projectiles", func(filename string, stream *kaitai.Stream) error {
		pro := p.NewPro()
		if err := pro.Read(stream, nil, pro); err != nil {
			return err
		}
		collection.LoadContextFromProjectile(filename, pro)
		return nil
	})
}

func processSpells(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	return processFiles(infFs, verbose, "SPL", "spells", func(filename string, stream *kaitai.Stream) error {
		spl := p.NewSpl()
		if err := spl.Read(stream, nil, spl); err != nil {
			return err
		}
		collection.LoadContextFromSpell(filename, spl)
		return nil
	})
}

func processStores(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	return processFiles(infFs, verbose, "STO", "stores", func(filename string, stream *kaitai.Stream) error {
		sto := p.NewSto()
		if err := sto.Read(stream, nil, sto); err != nil {
			return err
		}
		collection.LoadContextFromStore(filename, sto)
		return nil
	})
}

func processEffects(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	return processFiles(infFs, verbose, "EFF", "effects", func(filename string, stream *kaitai.Stream) error {
		eff := p.NewEff()
		if err := eff.Read(stream, nil, eff); err != nil {
			return err
		}
		collection.LoadContextFromEffect(filename, eff)
		return nil
	})
}

func process2daFiles(collection *text.TextCollection, infFs afero.Fs, verbose bool) error {
	// Parse SNDSLOT.IDS for CHARSND.2DA context
	var sndslotIds *p.Ids
	sndslot, err := infFs.Open("SNDSLOT.IDS")
	if err == nil {
		sndslotIds, _ = p.ParseIds(sndslot)
		sndslot.Close()
	}

	type twodaProcessor struct {
		filename string
		loadFunc func(*text.TextCollection, string, *p.TwoDA) error
	}

	processors := []twodaProcessor{
		{"25ECRED.2DA", (*text.TextCollection).LoadContextFrom25ECred2DA},
		{"25STWEAP.2DA", (*text.TextCollection).LoadContextFrom25StWeap2DA},
		{"7eyes.2DA", (*text.TextCollection).LoadContextFrom7Eyes2DA},
		{"BDSTWEAP.2DA", (*text.TextCollection).LoadContextFrom25StWeap2DA},
		{"CHARSND.2DA", func(c *text.TextCollection, f string, t *p.TwoDA) error {
			return c.LoadContextFromCharSnd2DA(f, t, sndslotIds)
		}},
		{"EFFTEXT.2DA", (*text.TextCollection).LoadContextFromEffText2DA},
		{"ENGINEST.2DA", (*text.TextCollection).LoadContextFromEngineSt2DA},
		{"MSCHOOL.2DA", (*text.TextCollection).LoadContextFromMSchool2DA},
		{"MSECTYPE.2DA", (*text.TextCollection).LoadContextFromMSecType2DA},
		{"TRACKING.2DA", (*text.TextCollection).LoadContextFromTracking2DA},
	}

	total := len(processors)
	processed := 0
	hasWarnings := false

	if verbose {
		fmt.Print("extracting context from 2DA files...")
	}

	for _, proc := range processors {
		file, err := infFs.Open(proc.filename)
		if err != nil {
			if verbose {
				if !hasWarnings {
					fmt.Println()
					hasWarnings = true
				}
				fmt.Printf("  warning: unable to open %s: %v. skipping...\n", proc.filename, err)
			}
			continue
		}

		twoda, err := p.ParseTwoDA(file)
		file.Close()
		if err != nil {
			if verbose {
				if !hasWarnings {
					fmt.Println()
					hasWarnings = true
				}
				fmt.Printf("  warning: unable to parse %s: %v. skipping...\n", proc.filename, err)
			}
			continue
		}

		proc.loadFunc(collection, proc.filename, twoda)
		processed++
	}

	if verbose {
		if processed == total {
			fmt.Printf(" done (%d files).\n", total)
		} else {
			fmt.Printf("done (%d/%d files).\n", processed, total)
		}
	}

	return nil
}

type TimestampEntry struct {
	Text      string
	Timestamp int64
}

func loadTimestamps(csvPath string) (map[uint32]TimestampEntry, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open timestamps CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("unable to parse timestamps CSV file: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("timestamps CSV file is empty")
	}

	// Find column indices from header
	header := records[0]
	idIdx, textIdx, timestampIdx := -1, -1, -1
	for i, col := range header {
		switch col {
		case "id":
			idIdx = i
		case "text":
			textIdx = i
		case "timestamp":
			timestampIdx = i
		}
	}

	if idIdx == -1 || textIdx == -1 || timestampIdx == -1 {
		return nil, fmt.Errorf("timestamps CSV file must have 'id', 'text', and 'timestamp' columns")
	}

	timestamps := make(map[uint32]TimestampEntry)
	for _, record := range records[1:] {
		if len(record) <= idIdx || len(record) <= textIdx || len(record) <= timestampIdx {
			continue
		}

		id, err := strconv.ParseUint(record[idIdx], 10, 32)
		if err != nil {
			continue
		}

		timestamp, _ := strconv.ParseInt(record[timestampIdx], 10, 64)
		timestamps[uint32(id)] = TimestampEntry{
			Text:      record[textIdx],
			Timestamp: timestamp,
		}
	}

	return timestamps, nil
}
