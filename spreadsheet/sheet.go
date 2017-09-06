// Copyright 2017 Baliance. All rights reserved.
//
// Use of this source code is governed by the terms of the Affero GNU General
// Public License version 3.0 as published by the Free Software Foundation and
// appearing in the file LICENSE included in the packaging of this file. A
// commercial license can be purchased by contacting sales@baliance.com.

package spreadsheet

import (
	"fmt"

	"baliance.com/gooxml"
	"baliance.com/gooxml/common"
	sml "baliance.com/gooxml/schema/schemas.openxmlformats.org/spreadsheetml"
)

// Sheet is a single sheet within a workbook.
type Sheet struct {
	w  *Workbook
	x  *sml.CT_Sheet
	ws *sml.Worksheet
}

// EnsureRow will return a row with a given row number, creating a new row if
// necessary.
func (s Sheet) EnsureRow(rowNum uint32) Row {
	// see if the row exists
	for _, r := range s.ws.SheetData.Row {
		if r.RAttr != nil && *r.RAttr == rowNum {
			return Row{s.w, r}
		}
	}
	// create a new row
	return s.AddNumberedRow(rowNum)
}

// AddNumberedRow adds a row with a given row number.  If you reuse a row number
// the resulting file will fail validation and fail to open in Office programs. Use
// EnsureRow instead which creates a new row or returns an existing row.
func (s Sheet) AddNumberedRow(rowNum uint32) Row {
	r := sml.NewCT_Row()
	r.RAttr = gooxml.Uint32(rowNum)
	s.ws.SheetData.Row = append(s.ws.SheetData.Row, r)
	return Row{s.w, r}
}

// AddRow adds a new row to a sheet.
func (s Sheet) AddRow() Row {
	maxRowID := uint32(0)
	// find the max row number
	for _, r := range s.ws.SheetData.Row {
		if r.RAttr != nil && *r.RAttr > maxRowID {
			maxRowID = *r.RAttr
		}
	}

	return s.AddNumberedRow(maxRowID + 1)
}

// Name returns the sheet name
func (s Sheet) Name() string {
	return s.x.NameAttr
}

// SetName sets the sheet name.
func (s Sheet) SetName(name string) {
	s.x.NameAttr = name
}

// Validate validates the sheet, returning an error if it is found to be invalid.
func (s Sheet) Validate() error {

	usedRows := map[uint32]struct{}{}
	for _, r := range s.ws.SheetData.Row {
		if r.RAttr != nil {
			if _, reusedRow := usedRows[*r.RAttr]; reusedRow {
				return fmt.Errorf("'%s' reused row %d", s.Name(), *r.RAttr)
			}
			usedRows[*r.RAttr] = struct{}{}
		}
		usedCells := map[string]struct{}{}
		for _, c := range r.C {
			if c.RAttr == nil {
				continue
			}
			if _, reusedCell := usedCells[*c.RAttr]; reusedCell {
				return fmt.Errorf("'%s' reused cell %s", s.Name(), *c.RAttr)
			}
			usedCells[*c.RAttr] = struct{}{}
		}
	}
	if err := s.x.Validate(); err != nil {
		return err
	}
	return s.ws.Validate()
}

// ValidateWithPath validates the sheet passing path informaton for a better
// error message
func (s Sheet) ValidateWithPath(path string) error {
	return s.x.ValidateWithPath(path)
}

// Rows returns all of the rows in a sheet.
func (s Sheet) Rows() []Row {
	ret := []Row{}
	for _, r := range s.ws.SheetData.Row {
		ret = append(ret, Row{s.w, r})
	}
	return ret
}

// SetDrawing sets the worksheet drawing.  A worksheet can have a reference to a
// single drawing, but the drawing can have many charts.
func (s Sheet) SetDrawing(d Drawing) {
	var rel common.Relationships
	for i, wks := range s.w.xws {
		if wks == s.ws {
			rel = s.w.xwsRels[i]
			break
		}
	}
	// add relationship from drawing to the sheet
	var drawingID string
	for i, dr := range d.wb.drawings {
		if dr == d.x {
			rel := rel.AddAutoRelationship(gooxml.DocTypeSpreadsheet, i+1, gooxml.DrawingType)
			drawingID = rel.ID()
			break
		}
	}
	s.ws.Drawing = sml.NewCT_Drawing()
	s.ws.Drawing.IdAttr = drawingID
}
