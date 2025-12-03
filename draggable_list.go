package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// DraggableList is a custom list widget that supports drag and drop to reorder items.
type DraggableList struct {
	widget.List
	OnReordered func(int, int)
	draggedItem int
	itemHeight  float32
}

// NewDraggableList creates a new draggable list.
func NewDraggableList(length func() int, createItem func() fyne.CanvasObject, updateItem func(widget.ListItemID, fyne.CanvasObject)) *DraggableList {
	list := &DraggableList{
		draggedItem: -1,
	}
	list.Length = length
	list.CreateItem = createItem
	list.UpdateItem = updateItem
	list.ExtendBaseWidget(list)

	// Calculate item height
	templateItem := createItem()
	list.itemHeight = templateItem.MinSize().Height

	return list
}

// Dragged is called when an item is dragged.
func (l *DraggableList) Dragged(e *fyne.DragEvent) {
	if l.itemHeight == 0 {
		return
	}

	if l.draggedItem == -1 {
		l.draggedItem = int(e.PointEvent.Position.Y / l.itemHeight)
	}

	newPos := int(e.PointEvent.Position.Y / l.itemHeight)

	if newPos > -1 && newPos < l.Length() && newPos != l.draggedItem {
		if l.OnReordered != nil {
			l.OnReordered(l.draggedItem, newPos)
		}
		l.draggedItem = newPos
	}
}

// DragEnd is called when a drag operation ends.
func (l *DraggableList) DragEnd() {
	l.draggedItem = -1
	l.Refresh()
}

// MouseIn is called when the mouse enters the widget.
func (l *DraggableList) MouseIn(*desktop.MouseEvent) {}

// MouseOut is called when the mouse leaves the widget.
func (l *DraggableList) MouseOut() {}

// MouseMoved is called when the mouse moves over the widget.
func (l *DraggableList) MouseMoved(*desktop.MouseEvent) {}
