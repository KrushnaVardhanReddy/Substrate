package sql

import (
	"fmt"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
)

func DiffSchemas(base, head *SQLSchema) *report.DiffReport {
	rep := &report.DiffReport{
		SubstrateVersion: "0.1.0",
		SchemaType:       report.SchemaTypeSQL,
		ComparedAt:       time.Now().UTC().Format(time.RFC3339),
		BreakingChanges:  []report.Change{},
		Warnings:         []report.Change{},
		SafeChanges:      []report.Change{},
	}

	if base == nil || head == nil {
		return rep
	}

	// Helper to add changes
	addChange := func(id, ruleID string, severity report.ChangeSeverity, path, desc string) {
		chg := report.Change{
			ID:          id,
			RuleID:      ruleID,
			Severity:    severity,
			Path:        path,
			Description: desc,
		}
		if severity == report.ChangeSeverityBreaking {
			rep.BreakingChanges = append(rep.BreakingChanges, chg)
			rep.Summary.BreakingCount++
		} else if severity == report.ChangeSeverityWarning {
			rep.Warnings = append(rep.Warnings, chg)
			rep.Summary.WarningCount++
		} else {
			rep.SafeChanges = append(rep.SafeChanges, chg)
			rep.Summary.SafeCount++
		}
		rep.Summary.TotalChanges++
	}

	// STEP 1 - Tables
	for tName, bTable := range base.Tables {
		if hTable, exists := head.Tables[tName]; !exists {
			addChange(
				fmt.Sprintf("chg_table_removed_%s", tName),
				"TABLE_REMOVED",
				report.ChangeSeverityBreaking,
				fmt.Sprintf("tables.%s", tName),
				fmt.Sprintf("Table '%s' was removed.", tName),
			)
		} else {
			diffColumns(bTable, hTable, addChange)
			diffConstraints(bTable, hTable, addChange)
			diffIndexes(bTable, hTable, addChange)
		}
	}

	for tName, hTable := range head.Tables {
		if _, exists := base.Tables[tName]; !exists {
			// Check RENAME heuristic for table
			renamedFrom := ""
			for bName, bTable := range base.Tables {
				if _, hExists := head.Tables[bName]; !hExists {
					score := levenshteinSimilarity(bName, tName)
					if score > 0.7 && sameColumns(bTable, hTable) {
						renamedFrom = bName
						break
					}
				}
			}

			if renamedFrom != "" {
				addChange(
					fmt.Sprintf("chg_table_renamed_%s", tName),
					"TABLE_RENAMED",
					report.ChangeSeverityBreaking,
					fmt.Sprintf("tables.%s", tName),
					fmt.Sprintf("Table '%s' was renamed to '%s'.", renamedFrom, tName),
				)
			} else {
				addChange(
					fmt.Sprintf("chg_table_added_%s", tName),
					"TABLE_ADDED",
					report.ChangeSeveritySafe,
					fmt.Sprintf("tables.%s", tName),
					fmt.Sprintf("Table '%s' was added.", tName),
				)
			}
		}
	}

	// STEP 5 - Views
	for vName, bView := range base.Views {
		if hView, exists := head.Views[vName]; !exists {
			addChange(
				fmt.Sprintf("chg_view_removed_%s", vName),
				"VIEW_REMOVED",
				report.ChangeSeverityBreaking,
				fmt.Sprintf("views.%s", vName),
				fmt.Sprintf("View '%s' was removed.", vName),
			)
		} else {
			// diff columns in view
			hCols := make(map[string]bool)
			for _, c := range hView.Columns {
				hCols[c] = true
			}
			var columnDiff []string
			for _, c := range bView.Columns {
				if !hCols[c] {
					columnDiff = append(columnDiff, c)
					addChange(
						fmt.Sprintf("chg_view_column_removed_%s_%s", vName, c),
						"VIEW_COLUMN_REMOVED",
						report.ChangeSeverityBreaking,
						fmt.Sprintf("views.%s.columns.%s", vName, c),
						fmt.Sprintf("Column '%s' was removed from view '%s'.", c, vName),
					)
				}
			}

			// VIEW_DEFINITION_CHANGED: same columns, different definition
			if len(columnDiff) == 0 && bView.Definition != "" && hView.Definition != "" {
				if bView.Definition != hView.Definition {
					addChange(
						fmt.Sprintf("chg_view_definition_changed_%s", vName),
						"VIEW_DEFINITION_CHANGED",
						report.ChangeSeverityWarning,
						fmt.Sprintf("views.%s", vName),
						fmt.Sprintf("View '%s' definition changed — same columns but different query.", vName),
					)
				}
			}
		}
	}
	for vName := range head.Views {
		if _, exists := base.Views[vName]; !exists {
			addChange(
				fmt.Sprintf("chg_view_added_%s", vName),
				"VIEW_ADDED",
				report.ChangeSeveritySafe,
				fmt.Sprintf("views.%s", vName),
				fmt.Sprintf("View '%s' was added.", vName),
			)
		}
	}

	// STEP 6 - Enums
	for eName, bEnum := range base.Enums {
		if hEnum, exists := head.Enums[eName]; !exists {
			addChange(
				fmt.Sprintf("chg_sql_enum_type_removed_%s", eName),
				"SQL_ENUM_TYPE_REMOVED",
				report.ChangeSeverityBreaking,
				fmt.Sprintf("enums.%s", eName),
				fmt.Sprintf("Enum type '%s' was removed.", eName),
			)
		} else {
			hVals := make(map[string]bool)
			for _, v := range hEnum.Values {
				hVals[v] = true
			}
			for _, v := range bEnum.Values {
				if !hVals[v] {
					addChange(
						fmt.Sprintf("chg_sql_enum_value_removed_%s_%s", eName, v),
						"SQL_ENUM_VALUE_REMOVED",
						report.ChangeSeverityBreaking,
						fmt.Sprintf("enums.%s.values.%s", eName, v),
						fmt.Sprintf("Enum value '%s' removed from type '%s'.", v, eName),
					)
				}
			}

			bVals := make(map[string]bool)
			for _, v := range bEnum.Values {
				bVals[v] = true
			}
			for _, v := range hEnum.Values {
				if !bVals[v] {
					addChange(
						fmt.Sprintf("chg_sql_enum_value_added_%s_%s", eName, v),
						"SQL_ENUM_VALUE_ADDED",
						report.ChangeSeverityWarning,
						fmt.Sprintf("enums.%s.values.%s", eName, v),
						fmt.Sprintf("Enum value '%s' added to type '%s'.", v, eName),
					)
				}
			}
		}
	}

	if rep.Summary.BreakingCount > 0 {
		rep.Summary.OverallSeverity = report.SeverityBreaking
	} else if rep.Summary.WarningCount > 0 {
		rep.Summary.OverallSeverity = report.SeverityWarning
	} else if rep.Summary.TotalChanges > 0 {
		rep.Summary.OverallSeverity = report.SeveritySafe
	} else {
		rep.Summary.OverallSeverity = report.SeverityNoChanges
	}

	return rep
}

func diffColumns(bTable, hTable *Table, addChange func(id, ruleID string, severity report.ChangeSeverity, path, desc string)) {
	tName := bTable.Name
	for cName, bCol := range bTable.Columns {
		hCol, exists := hTable.Columns[cName]
		if !exists {
			// Check RENAME heuristic
			renamedTo := ""
			for hName, hc := range hTable.Columns {
				if _, bExists := bTable.Columns[hName]; !bExists {
					score := levenshteinSimilarity(cName, hName)
					if score > 0.7 && bCol.DataType == hc.DataType {
						renamedTo = hName
						break
					}
				}
			}

			if renamedTo != "" {
				addChange(
					fmt.Sprintf("chg_column_renamed_%s_%s", tName, cName),
					"COLUMN_RENAMED",
					report.ChangeSeverityBreaking,
					fmt.Sprintf("tables.%s.columns.%s", tName, cName),
					fmt.Sprintf("Column '%s' in table '%s' was renamed to '%s'.", cName, tName, renamedTo),
				)
			} else {
				addChange(
					fmt.Sprintf("chg_column_removed_%s_%s", tName, cName),
					"COLUMN_REMOVED",
					report.ChangeSeverityBreaking,
					fmt.Sprintf("tables.%s.columns.%s", tName, cName),
					fmt.Sprintf("Column '%s' was removed from table '%s'.", cName, tName),
				)
			}
		} else {
			diffSingleColumn(bTable.Name, bCol, hCol, addChange)
		}
	}

	for cName, hCol := range hTable.Columns {
		if _, exists := bTable.Columns[cName]; !exists {
			// skip if it was a rename
			wasRenamed := false
			for bName, bc := range bTable.Columns {
				if _, hExists := hTable.Columns[bName]; !hExists {
					if levenshteinSimilarity(bName, cName) > 0.7 && bc.DataType == hCol.DataType {
						wasRenamed = true
						break
					}
				}
			}
			if wasRenamed {
				continue
			}

			if hCol.NotNull && hCol.Default == nil {
				addChange(
					fmt.Sprintf("chg_column_added_not_null_no_default_%s_%s", tName, cName),
					"COLUMN_ADDED_NOT_NULL_NO_DEFAULT",
					report.ChangeSeverityBreaking,
					fmt.Sprintf("tables.%s.columns.%s", tName, cName),
					fmt.Sprintf("Column '%s' added as NOT NULL without a DEFAULT — existing INSERTs will fail.", cName),
				)
			} else {
				addChange(
					fmt.Sprintf("chg_column_added_nullable_%s_%s", tName, cName),
					"COLUMN_ADDED_NULLABLE",
					report.ChangeSeveritySafe,
					fmt.Sprintf("tables.%s.columns.%s", tName, cName),
					fmt.Sprintf("Column '%s' added to table '%s'.", cName, tName),
				)
			}
		}
	}
}

func diffSingleColumn(tName string, bCol, hCol *Column, addChange func(id, ruleID string, severity report.ChangeSeverity, path, desc string)) {
	cName := bCol.Name

	lengthChanged := (bCol.Length == nil && hCol.Length != nil) || (bCol.Length != nil && hCol.Length == nil) || (bCol.Length != nil && hCol.Length != nil && *bCol.Length != *hCol.Length)
	precisionChanged := (bCol.Precision == nil && hCol.Precision != nil) || (bCol.Precision != nil && hCol.Precision == nil) || (bCol.Precision != nil && hCol.Precision != nil && *bCol.Precision != *hCol.Precision)

	if bCol.DataType != hCol.DataType || lengthChanged || precisionChanged {
		if isWidening(bCol, hCol) {
			addChange(
				fmt.Sprintf("chg_column_type_widened_%s_%s", tName, cName),
				"COLUMN_TYPE_WIDENED",
				report.ChangeSeverityWarning,
				fmt.Sprintf("tables.%s.columns.%s", tName, cName),
				fmt.Sprintf("Column '%s' type widened from '%s' to '%s'.", cName, typeString(bCol), typeString(hCol)),
			)
		} else {
			addChange(
				fmt.Sprintf("chg_column_type_changed_%s_%s", tName, cName),
				"COLUMN_TYPE_CHANGED",
				report.ChangeSeverityBreaking,
				fmt.Sprintf("tables.%s.columns.%s", tName, cName),
				fmt.Sprintf("Column '%s' type changed from '%s' to '%s'.", cName, typeString(bCol), typeString(hCol)),
			)
		}
	}

	if !bCol.NotNull && hCol.NotNull {
		addChange(
			fmt.Sprintf("chg_column_made_not_null_%s_%s", tName, cName),
			"COLUMN_MADE_NOT_NULL",
			report.ChangeSeverityBreaking,
			fmt.Sprintf("tables.%s.columns.%s", tName, cName),
			fmt.Sprintf("Column '%s' changed from nullable to NOT NULL — existing NULLs will be rejected.", cName),
		)
	} else if bCol.NotNull && !hCol.NotNull {
		addChange(
			fmt.Sprintf("chg_column_nullable_changed_%s_%s", tName, cName),
			"COLUMN_NULLABLE_CHANGED",
			report.ChangeSeveritySafe,
			fmt.Sprintf("tables.%s.columns.%s", tName, cName),
			fmt.Sprintf("Column '%s' changed from NOT NULL to nullable.", cName),
		)
	}

	bDef := bCol.Default
	hDef := hCol.Default

	if bDef != nil && hDef == nil {
		if hCol.NotNull {
			addChange(
				fmt.Sprintf("chg_column_default_removed_%s_%s", tName, cName),
				"COLUMN_DEFAULT_REMOVED",
				report.ChangeSeverityBreaking,
				fmt.Sprintf("tables.%s.columns.%s", tName, cName),
				fmt.Sprintf("DEFAULT removed from NOT NULL column '%s' in table '%s'.", cName, tName),
			)
		} else {
			addChange(
				fmt.Sprintf("chg_column_default_changed_%s_%s", tName, cName),
				"COLUMN_DEFAULT_CHANGED",
				report.ChangeSeverityWarning,
				fmt.Sprintf("tables.%s.columns.%s", tName, cName),
				fmt.Sprintf("DEFAULT removed from column '%s' in table '%s'.", cName, tName),
			)
		}
	} else if bDef != nil && hDef != nil && *bDef != *hDef {
		addChange(
			fmt.Sprintf("chg_column_default_changed_%s_%s", tName, cName),
			"COLUMN_DEFAULT_CHANGED",
			report.ChangeSeverityWarning,
			fmt.Sprintf("tables.%s.columns.%s", tName, cName),
			fmt.Sprintf("DEFAULT changed for column '%s' in table '%s'.", cName, tName),
		)
	} else if bDef == nil && hDef != nil {
		addChange(
			fmt.Sprintf("chg_column_default_changed_%s_%s", tName, cName),
			"COLUMN_DEFAULT_CHANGED",
			report.ChangeSeverityWarning,
			fmt.Sprintf("tables.%s.columns.%s", tName, cName),
			fmt.Sprintf("DEFAULT added to column '%s' in table '%s'.", cName, tName),
		)
	}
}

func isWidening(bCol, hCol *Column) bool {
	bt, ht := bCol.DataType, hCol.DataType
	if bt == "smallint" && (ht == "integer" || ht == "bigint") {
		return true
	}
	if bt == "integer" && ht == "bigint" {
		return true
	}
	if bt == "real" && ht == "double precision" {
		return true
	}
	if bt == "varchar" && ht == "varchar" {
		if bCol.Length != nil && hCol.Length != nil && *hCol.Length > *bCol.Length {
			return true
		}
		if bCol.Length != nil && hCol.Length == nil { // widening to unlimited
			return true
		}
	}
	if (bt == "numeric" || bt == "decimal") && (ht == "numeric" || ht == "decimal") {
		if bCol.Precision != nil && hCol.Precision != nil && *hCol.Precision > *bCol.Precision &&
			(bCol.Scale == hCol.Scale || (bCol.Scale == nil && hCol.Scale == nil) || (bCol.Scale != nil && hCol.Scale != nil && *bCol.Scale == *hCol.Scale)) {
			return true
		}
		if bCol.Precision != nil && hCol.Precision == nil {
			return true
		}
	}
	return false
}

func typeString(c *Column) string {
	if c.DataType == "varchar" || c.DataType == "char" {
		if c.Length != nil {
			return fmt.Sprintf("%s(%d)", c.DataType, *c.Length)
		}
	}
	if c.DataType == "numeric" || c.DataType == "decimal" {
		if c.Precision != nil && c.Scale != nil {
			return fmt.Sprintf("%s(%d,%d)", c.DataType, *c.Precision, *c.Scale)
		} else if c.Precision != nil {
			return fmt.Sprintf("%s(%d)", c.DataType, *c.Precision)
		}
	}
	return c.DataType
}

func diffConstraints(bTable, hTable *Table, addChange func(id, ruleID string, severity report.ChangeSeverity, path, desc string)) {
	tName := bTable.Name
	for cName, bCon := range bTable.Constraints {
		if _, exists := hTable.Constraints[cName]; !exists {
			if bCon.Type == "PRIMARY_KEY" {
				addChange(
					fmt.Sprintf("chg_primary_key_changed_%s", tName),
					"PRIMARY_KEY_CHANGED",
					report.ChangeSeverityBreaking,
					fmt.Sprintf("tables.%s", tName),
					fmt.Sprintf("PRIMARY KEY constraint '%s' removed or changed.", cName),
				)
			} else {
				addChange(
					fmt.Sprintf("chg_constraint_removed_%s_%s", tName, cName),
					"CONSTRAINT_REMOVED",
					report.ChangeSeverityWarning,
					fmt.Sprintf("tables.%s", tName),
					fmt.Sprintf("Constraint '%s' was removed from table '%s'.", cName, tName),
				)
			}
		}
	}

	for cName, hCon := range hTable.Constraints {
		if _, exists := bTable.Constraints[cName]; !exists {
			if hCon.Type == "PRIMARY_KEY" {
				addChange(
					fmt.Sprintf("chg_primary_key_changed_%s", tName),
					"PRIMARY_KEY_CHANGED",
					report.ChangeSeverityBreaking,
					fmt.Sprintf("tables.%s", tName),
					fmt.Sprintf("PRIMARY KEY constraint '%s' added or changed.", cName),
				)
			} else if hCon.Type == "UNIQUE" {
				addChange(
					fmt.Sprintf("chg_constraint_added_unique_%s_%s", tName, cName),
					"CONSTRAINT_ADDED_UNIQUE",
					report.ChangeSeverityBreaking,
					fmt.Sprintf("tables.%s", tName),
					fmt.Sprintf("UNIQUE constraint '%s' added to table '%s'.", cName, tName),
				)
			} else if hCon.Type == "CHECK" {
				addChange(
					fmt.Sprintf("chg_constraint_added_check_%s_%s", tName, cName),
					"CONSTRAINT_ADDED_CHECK",
					report.ChangeSeverityBreaking,
					fmt.Sprintf("tables.%s", tName),
					fmt.Sprintf("CHECK constraint '%s' added to table '%s'.", cName, tName),
				)
			} else if hCon.Type == "FOREIGN_KEY" {
				addChange(
					fmt.Sprintf("chg_foreign_key_added_%s_%s", tName, cName),
					"FOREIGN_KEY_ADDED",
					report.ChangeSeverityBreaking,
					fmt.Sprintf("tables.%s", tName),
					fmt.Sprintf("FOREIGN KEY '%s' added to table '%s'.", cName, tName),
				)
			}
		}
	}
}

func diffIndexes(bTable, hTable *Table, addChange func(id, ruleID string, severity report.ChangeSeverity, path, desc string)) {
	tName := bTable.Name
	for iName := range bTable.Indexes {
		if _, exists := hTable.Indexes[iName]; !exists {
			addChange(
				fmt.Sprintf("chg_index_removed_%s_%s", tName, iName),
				"INDEX_REMOVED",
				report.ChangeSeverityWarning,
				fmt.Sprintf("tables.%s", tName),
				fmt.Sprintf("Index '%s' removed from table '%s'.", iName, tName),
			)
		}
	}

	for iName := range hTable.Indexes {
		if _, exists := bTable.Indexes[iName]; !exists {
			addChange(
				fmt.Sprintf("chg_index_added_%s_%s", tName, iName),
				"INDEX_ADDED",
				report.ChangeSeveritySafe,
				fmt.Sprintf("tables.%s", tName),
				fmt.Sprintf("Index '%s' added to table '%s'.", iName, tName),
			)
		}
	}
}

func sameColumns(t1, t2 *Table) bool {
	if len(t1.Columns) != len(t2.Columns) {
		return false
	}
	for cName := range t1.Columns {
		if _, exists := t2.Columns[cName]; !exists {
			return false
		}
	}
	return true
}

func levenshteinSimilarity(s1, s2 string) float64 {
	d := levenshteinDistance(s1, s2)
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(d)/float64(maxLen)
}

func levenshteinDistance(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	n, m := len(r1), len(r2)
	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}

	d := make([][]int, n+1)
	for i := range d {
		d[i] = make([]int, m+1)
	}

	for i := 0; i <= n; i++ {
		d[i][0] = i
	}
	for j := 0; j <= m; j++ {
		d[0][j] = j
	}

	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			cost := 1
			if r1[i-1] == r2[j-1] {
				cost = 0
			}
			min := d[i-1][j] + 1
			if d[i][j-1]+1 < min {
				min = d[i][j-1] + 1
			}
			if d[i-1][j-1]+cost < min {
				min = d[i-1][j-1] + cost
			}
			d[i][j] = min
		}
	}
	return d[n][m]
}
