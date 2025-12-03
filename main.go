package main

import (
	"fmt"
	"os"
	"regexp"

	"grub-configurator/grub"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Grub Configurator")

	myWindow.Resize(fyne.NewSize(800, 600))

	// Grub Config Tab
	grubConfigTab := createGrubConfigTab(myWindow)

	// Boot Order Tab
	bootOrderTab := createBootOrderTab(myWindow)

	tabs := container.NewAppTabs(
		container.NewTabItem("Grub Config", grubConfigTab),
		container.NewTabItem("Boot Order", bootOrderTab),
	)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}

func createGrubConfigTab(myWindow fyne.Window) fyne.CanvasObject {
	// Read the GRUB config file
	grubBytes, err := os.ReadFile(grub.GrubConfigPath)
	if err != nil {
		// If the file doesn't exist, create it with some default content
		if os.IsNotExist(err) {
			defaultContent := "GRUB_DEFAULT=0\nGRUB_TIMEOUT=5\nGRUB_DISTRIBUTOR=`lsb_release -i -s 2> /dev/null || echo Debian`\nGRUB_CMDLINE_LINUX_DEFAULT=\"quiet splash\"\nGRUB_CMDLINE_LINUX=\"\"\n"
			if err := grub.WriteGrubConfig(defaultContent); err != nil {
				dialog.ShowError(fmt.Errorf("failed to create default grub config: %w", err), myWindow)
				return widget.NewLabel("Failed to create default grub config")
			}
			if err := grub.DisableOsProber(); err != nil {
				dialog.ShowError(fmt.Errorf("failed to disable os prober after creating default config: %w", err), myWindow)
				return widget.NewLabel("Failed to disable os prober")
			}
			grubBytes = []byte(defaultContent)
		} else {
			dialog.ShowError(fmt.Errorf("failed to read grub config: %w", err), myWindow)
			return widget.NewLabel("Failed to read grub config")
		}
	}

	// Create a multi-line text input
	grubInput := widget.NewMultiLineEntry()
	grubInput.SetText(string(grubBytes))

	// Create a save button
	saveButton := widget.NewButton("Save", func() {
		dialog.ShowConfirm("Save Grub Config", "Are you sure you want to save the changes to the grub config?", func(ok bool) {
			if !ok {
				return
			}
			newContent := grubInput.Text
			// Write the changes back to the file
			if err := grub.WriteGrubConfig(newContent); err != nil {
				dialog.ShowError(fmt.Errorf("failed to write grub config: %w", err), myWindow)
				return
			}
			dialog.ShowInformation("Success", "Grub config saved successfully!", myWindow)
		}, myWindow)
	})

	// Create an update grub button
	updateButton := widget.NewButton("Update Grub", func() {
		progress := dialog.NewProgressInfinite("Running update-grub", "Please wait...", myWindow)
		progress.Show()
		output, err := grub.RunGrubMkconfig()
		progress.Hide()
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to update grub: %w\n%s", err, output), myWindow)
			return
		}
		dialog.ShowInformation("Success", "Grub updated successfully!", myWindow)
	})

	buttons := container.NewHBox(saveButton, updateButton)
	return container.NewBorder(nil, buttons, nil, nil, container.NewScroll(grubInput))
}

func createBootOrderTab(myWindow fyne.Window) fyne.CanvasObject {
	// Read the grub.cfg file
	menuEntries, err := grub.ParseGrubCfg()
	if err != nil {
		return widget.NewLabel(fmt.Sprintf("Failed to parse grub.cfg: %v", err))
	}

	// Create a draggable list to display the menu entries
	list := NewDraggableList(
		func() int {
			return len(menuEntries)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(menuEntries[i].Title)
		},
	)

	var selected widget.ListItemID = -1
	list.OnSelected = func(id widget.ListItemID) {
		selected = id
	}

	list.OnReordered = func(from, to int) {
		menuEntries[from], menuEntries[to] = menuEntries[to], menuEntries[from]
		list.Refresh()
	}

	// Create buttons to move items
	moveUpButton := widget.NewButton("Move Up", func() {
		if selected > 0 {
			menuEntries[selected], menuEntries[selected-1] = menuEntries[selected-1], menuEntries[selected]
			list.Refresh()
			list.Select(selected - 1)
		}
	})

	moveDownButton := widget.NewButton("Move Down", func() {
		if selected != -1 && selected < len(menuEntries)-1 {
			menuEntries[selected], menuEntries[selected+1] = menuEntries[selected+1], menuEntries[selected]
			list.Refresh()
			list.Select(selected + 1)
		}
	})

	// Create remove button
	removeButton := widget.NewButton("Remove", func() {
		if selected == -1 {
			return
		}
		dialog.ShowConfirm("Delete entry", "Are you sure you want to delete this entry?", func(ok bool) {
			if ok {
				menuEntries = append(menuEntries[:selected], menuEntries[selected+1:]...)
				list.Refresh()
				list.UnselectAll()
				selected = -1
			}
		}, myWindow)
	})

	// Create a save button
	saveButton := widget.NewButton("Save and Update Grub", func() {
		dialog.ShowConfirm("Save and Update Grub", "Are you sure you want to save the changes to the boot order?", func(ok bool) {
			if !ok {
				return
			}
			progress := dialog.NewProgressInfinite("Saving and updating grub", "Please wait...", myWindow)
			progress.Show()

			// Write the new boot order to the custom script
			if err := grub.WriteCustomProxyScript(menuEntries); err != nil {
				progress.Hide()
				dialog.ShowError(fmt.Errorf("failed to write custom proxy script: %w", err), myWindow)
				return
			}

			// ask if user wants to disable OS prober
			if err := grub.DisableOsProber(); err != nil {
				progress.Hide()
				dialog.ShowError(fmt.Errorf("failed to disable os prober: %w", err), myWindow)
				return
			}

			// Run update-grub
			output, err := grub.RunGrubMkconfig()
			progress.Hide()
			if err != nil {
				dialog.ShowError(fmt.Errorf("failed to update grub: %w\n%s", err, output), myWindow)
				return
			}
			dialog.ShowInformation("Success", "Grub updated successfully!", myWindow)
		}, myWindow)
	})

	// Create a rename button
	renameButton := widget.NewButton("Rename", func() {
		if selected != -1 {
			entry := menuEntries[selected]
			newTitle := widget.NewEntry()
			newTitle.SetText(entry.Title)

			dialog.ShowForm("Rename", "Rename the selected boot entry", "Rename", []*widget.FormItem{
				widget.NewFormItem("New Name", newTitle),
			}, func(ok bool) {
				if ok {
					oldTitle := menuEntries[selected].Title
					newTitleText := newTitle.Text
					menuEntries[selected].Title = newTitleText
					// Use a regular expression to safely replace the title in the content
					re := regexp.MustCompile(fmt.Sprintf(`(menuentry ')(%s)(')`, regexp.QuoteMeta(oldTitle)))
					menuEntries[selected].Content = re.ReplaceAllString(menuEntries[selected].Content, fmt.Sprintf("${1}%s${3}", newTitleText))
					list.Refresh()
				}
			}, myWindow)
		}
	})

	// Add the rename button to the existing buttons layout
	buttons := container.NewVBox(moveUpButton, moveDownButton, renameButton, removeButton, saveButton)
	return container.NewBorder(nil, buttons, nil, nil, list)
}
