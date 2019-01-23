package table

import (
	"fmt"
	"github.com/attic-labs/noms/go/types"
	"github.com/liquidata-inc/ld/dolt/go/libraries/doltcore"
	"github.com/liquidata-inc/ld/dolt/go/libraries/doltcore/schema"
	"github.com/liquidata-inc/ld/dolt/go/libraries/utils/pantoerr"
)

var IdentityConverter = &RowConverter{nil, true, nil}

type RowConverter struct {
	*schema.FieldMapping
	IdentityConverter bool
	convFuncs         []doltcore.ConvFunc
}

func NewRowConverter(mapping *schema.FieldMapping) (*RowConverter, error) {
	if mapping.SrcSch == nil || mapping.DestSch == nil || len(mapping.DestToSrc) == 0 {
		panic("Invalid oldNameToSchema2Name cannot be used for conversion")
	}

	if !isNecessary(mapping.SrcSch, mapping.DestSch, mapping.DestToSrc) {
		return &RowConverter{nil, true, nil}, nil
	}

	convFuncs := make([]doltcore.ConvFunc, mapping.DestSch.NumFields())
	for dstIdx, srcIdx := range mapping.DestToSrc {
		if srcIdx != -1 {
			destFld := mapping.DestSch.GetField(dstIdx)
			srcFld := mapping.SrcSch.GetField(srcIdx)

			convFuncs[dstIdx] = doltcore.GetConvFunc(srcFld.NomsKind(), destFld.NomsKind())

			if convFuncs[dstIdx] == nil {
				return nil, fmt.Errorf("Unsupported conversion from type %s to %s", srcFld.KindString(), destFld.KindString())
			}
		}
	}

	return &RowConverter{mapping, false, convFuncs}, nil
}

func (rc *RowConverter) Convert(inRow *Row) (*RowData, error) {

	rowData := inRow.CurrData()
	return rc.ConvertRowData(rowData)
}

func (rc *RowConverter) ConvertRowData(rowData *RowData) (*RowData, error) {
	if rc.IdentityConverter {
		return rowData, nil
	}

	destFieldCount := rc.DestSch.NumFields()
	fieldVals := make([]types.Value, destFieldCount)

	unexpectedErr := NewBadRow(NewRow(rowData), "Unexpected Error")
	err := pantoerr.PanicToErrorInstance(unexpectedErr, func() error {
		for i := 0; i < destFieldCount; i++ {
			srcIdx := rc.DestToSrc[i]
			if srcIdx == -1 {
				continue
			}

			val, _ := rowData.GetField(srcIdx)
			mappedVal, err := rc.convFuncs[i](val)

			if err != nil {
				return NewBadRow(NewRow(rowData), err.Error())
			}

			fieldVals[i] = mappedVal
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return RowDataFromValues(rc.DestSch, fieldVals), nil
}

func isNecessary(srcSch, destSch *schema.Schema, destToSrc []int) bool {
	if len(destToSrc) != srcSch.NumFields() || len(destToSrc) != destSch.NumFields() {
		return true
	}

	for i := 0; i < len(destToSrc); i++ {
		if i != destToSrc[i] {
			return true
		}

		if srcSch.GetField(i).NomsKind() != destSch.GetField(i).NomsKind() {
			return true
		}
	}

	srcHasPK := srcSch.NumConstraintsOfType(schema.PrimaryKey) != 0
	destHasPK := destSch.NumConstraintsOfType(schema.PrimaryKey) != 0

	if srcHasPK != destHasPK {
		return true
	}

	if destHasPK && srcSch.GetPKIndex() != destToSrc[destSch.GetPKIndex()] {
		return true
	}

	return false
}