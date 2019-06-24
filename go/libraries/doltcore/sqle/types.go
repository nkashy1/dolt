package sqle

import (
	"fmt"
	"github.com/attic-labs/noms/go/types"
	"github.com/google/uuid"
	"github.com/src-d/go-mysql-server/sql"
)

func nomsTypeToSqlType(kind types.NomsKind) sql.Type {
	switch kind {
	case types.BoolKind:
		return sql.Boolean
	case types.FloatKind:
		return sql.Float64
	case types.StringKind:
		return sql.Text
	case types.UUIDKind:
		// TODO: make an actual uuid
		return sql.Text
	case types.IntKind:
		return sql.Int64
	case types.UintKind:
		return sql.Uint64
	default:
		panic(fmt.Sprintf("Unexpected kind %v", kind))
	}
}

func SqlTypeToNomsKind(t sql.Type) types.NomsKind {
	switch t {
	case sql.Boolean:
		return types.BoolKind
	case sql.Float64:
		return types.FloatKind
	case sql.Text:
		// TODO: handle UUIDs
		return types.StringKind
	case sql.Int64:
		return types.IntKind
	case sql.Uint64:
		return types.UintKind
	default:
		panic(fmt.Sprintf("Unexpected type %v", t))
	}
}

func nomsValToSqlVal(val types.Value) interface{} {
	switch val.Kind() {
	case types.BoolKind:
		return convertBool(val.(types.Bool))
	case types.FloatKind:
		return convertFloat(val.(types.Float))
	case types.StringKind:
		return convertString(val.(types.String))
	case types.UUIDKind:
		return convertUUID(val.(types.UUID))
	case types.IntKind:
		return convertInt(val.(types.Int))
	case types.UintKind:
		return convertUint(val.(types.Uint))
	default:
		panic(fmt.Sprintf("Unexpected kind %v", val.Kind()))
	}
}

func SqlValToNomsVal(val interface{}) types.Value {
	if val == nil {
		return nil
	}

	switch e := val.(type) {
	case int64:
		return types.Int(e)
	case float64:
		return types.Float(e)
	case string:
		if u, err := uuid.Parse(e); err == nil {
			return types.UUID(u)
		}
		return types.String(e)
	case uint64:
		return types.Uint(e)
	case bool:
		return types.Bool(e)
	default:
		panic(fmt.Sprintf("Unexpected type %v", val))
	}
}

func convertUUID(u types.UUID) interface{} {
	return u.String()
}

func convertUint(i types.Uint) interface{} {
	return uint64(i)
}

func convertInt(i types.Int) interface{} {
	return int64(i)
}

func convertString(i types.String) interface{} {
	return string(i)
}

func convertFloat(f types.Float) interface{} {
	return float64(f)
}

func convertBool(b types.Bool) interface{} {
	return bool(b)
}