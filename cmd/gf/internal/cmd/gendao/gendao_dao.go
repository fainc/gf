// Copyright GoFrame gf Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package gendao

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gview"
	"github.com/gogf/gf/v2/text/gstr"

	"github.com/gogf/gf/cmd/gf/v2/internal/consts"
	"github.com/gogf/gf/cmd/gf/v2/internal/utility/mlog"
	"github.com/gogf/gf/cmd/gf/v2/internal/utility/utils"
)

func generateDao(ctx context.Context, in CGenDaoInternalInput) {
	var (
		dirPathDao         = gfile.Join(in.Path, in.DaoPath)
		dirPathDaoInternal = gfile.Join(dirPathDao, "internal")
	)
	in.genItems.AppendDirPath(dirPathDao)
	for i := 0; i < len(in.TableNames); i++ {
		var (
			realTableName = in.TableNames[i]
			newTableName  = in.NewTableNames[i]
		)
		generateDaoSingle(ctx, generateDaoSingleInput{
			CGenDaoInternalInput: in,
			TableName:            realTableName,
			NewTableName:         newTableName,
			DirPathDao:           dirPathDao,
			DirPathDaoInternal:   dirPathDaoInternal,
			IsSharding:           in.ShardingTableSet.Contains(newTableName),
		})
	}
}

type generateDaoSingleInput struct {
	CGenDaoInternalInput
	// TableName specifies the table name of the table.
	TableName string
	// NewTableName specifies the prefix-stripped or custom edited name of the table.
	NewTableName       string
	DirPathDao         string
	DirPathDaoInternal string
	IsSharding         bool
}

// generateDaoSingle generates the dao and model content of given table.
func generateDaoSingle(ctx context.Context, in generateDaoSingleInput) {
	// Generating table data preparing.
	fieldMap, err := in.DB.TableFields(ctx, in.TableName)
	if err != nil {
		mlog.Fatalf(`fetching tables fields failed for table "%s": %+v`, in.TableName, err)
	}
	var (
		tableNameCamelCase      = formatFieldName(in.NewTableName, FieldNameCaseCamel)
		tableNameCamelLowerCase = formatFieldName(in.NewTableName, FieldNameCaseCamelLower)
		tableNameSnakeCase      = gstr.CaseSnake(in.NewTableName)
		importPrefix            = in.ImportPrefix
	)
	if importPrefix == "" {
		importPrefix = utils.GetImportPath(gfile.Join(in.Path, in.DaoPath))
	} else {
		importPrefix = gstr.Join(g.SliceStr{importPrefix, in.DaoPath}, "/")
	}

	fileName := gstr.Trim(tableNameSnakeCase, "-_.")
	if len(fileName) > 5 && fileName[len(fileName)-5:] == "_test" {
		// Add suffix to avoid the table name which contains "_test",
		// which would make the go file a testing file.
		fileName += "_table"
	}

	// dao - index
	generateDaoIndex(generateDaoIndexInput{
		generateDaoSingleInput:  in,
		TableNameCamelCase:      tableNameCamelCase,
		TableNameCamelLowerCase: tableNameCamelLowerCase,
		ImportPrefix:            importPrefix,
		FileName:                fileName,
	})

	// dao - internal
	generateDaoInternal(generateDaoInternalInput{
		generateDaoSingleInput:  in,
		TableNameCamelCase:      tableNameCamelCase,
		TableNameCamelLowerCase: tableNameCamelLowerCase,
		ImportPrefix:            importPrefix,
		FileName:                fileName,
		FieldMap:                fieldMap,
	})
}

type generateDaoIndexInput struct {
	generateDaoSingleInput
	TableNameCamelCase      string
	TableNameCamelLowerCase string
	ImportPrefix            string
	FileName                string
}

func generateDaoIndex(in generateDaoIndexInput) {
	path := filepath.FromSlash(gfile.Join(in.DirPathDao, in.FileName+".go"))
	// It should add path to result slice whenever it would generate the path file or not.
	in.genItems.AppendGeneratedFilePath(path)
	if in.OverwriteDao || !gfile.Exists(path) {
		var (
			ctx        = context.Background()
			tplContent = getTemplateFromPathOrDefault(
				in.TplDaoIndexPath, consts.TemplateGenDaoIndexContent,
			)
		)
		tplView.ClearAssigns()
		tplView.Assigns(gview.Params{
			tplVarTableSharding:           in.IsSharding,
			tplVarImportPrefix:            in.ImportPrefix,
			tplVarTableName:               in.TableName,
			tplVarTableNameCamelCase:      in.TableNameCamelCase,
			tplVarTableNameCamelLowerCase: in.TableNameCamelLowerCase,
			tplVarPackageName:             filepath.Base(in.DaoPath),
		})
		indexContent, err := tplView.ParseContent(ctx, tplContent)
		if err != nil {
			mlog.Fatalf("parsing template content failed: %v", err)
		}
		if err = gfile.PutContents(path, strings.TrimSpace(indexContent)); err != nil {
			mlog.Fatalf("writing content to '%s' failed: %v", path, err)
		} else {
			utils.GoFmt(path)
			mlog.Print("generated:", gfile.RealPath(path))
		}
	}
}

type generateDaoInternalInput struct {
	generateDaoSingleInput
	TableNameCamelCase      string
	TableNameCamelLowerCase string
	ImportPrefix            string
	FileName                string
	FieldMap                map[string]*gdb.TableField
}

func generateDaoInternal(in generateDaoInternalInput) {
	var (
		ctx                    = context.Background()
		removeFieldPrefixArray = gstr.SplitAndTrim(in.RemoveFieldPrefix, ",")
		tplContent             = getTemplateFromPathOrDefault(
			in.TplDaoInternalPath, consts.TemplateGenDaoInternalContent,
		)
	)
	tplView.ClearAssigns()
	tplView.Assigns(gview.Params{
		tplVarImportPrefix:            in.ImportPrefix,
		tplVarTableName:               in.TableName,
		tplVarGroupName:               in.Group,
		tplVarTableNameCamelCase:      in.TableNameCamelCase,
		tplVarTableNameCamelLowerCase: in.TableNameCamelLowerCase,
		tplVarColumnDefine:            gstr.Trim(generateColumnDefinitionForDao(in.FieldMap, removeFieldPrefixArray)),
		tplVarColumnNames:             gstr.Trim(generateColumnNamesForDao(in.FieldMap, removeFieldPrefixArray)),
	})
	assignDefaultVar(tplView, in.CGenDaoInternalInput)
	modelContent, err := tplView.ParseContent(ctx, tplContent)
	if err != nil {
		mlog.Fatalf("parsing template content failed: %v", err)
	}
	path := filepath.FromSlash(gfile.Join(in.DirPathDaoInternal, in.FileName+".go"))
	in.genItems.AppendGeneratedFilePath(path)
	if err := gfile.PutContents(path, strings.TrimSpace(modelContent)); err != nil {
		mlog.Fatalf("writing content to '%s' failed: %v", path, err)
	} else {
		utils.GoFmt(path)
		mlog.Print("generated:", gfile.RealPath(path))
	}
}

// generateColumnNamesForDao generates and returns the column names assignment content of column struct
// for specified table.
func generateColumnNamesForDao(fieldMap map[string]*gdb.TableField, removeFieldPrefixArray []string) string {
	var (
		buffer = bytes.NewBuffer(nil)
		array  = make([][]string, len(fieldMap))
		names  = sortFieldKeyForDao(fieldMap)
	)

	for index, name := range names {
		field := fieldMap[name]

		newFiledName := field.Name
		for _, v := range removeFieldPrefixArray {
			newFiledName = gstr.TrimLeftStr(newFiledName, v, 1)
		}

		array[index] = []string{
			"            #" + formatFieldName(newFiledName, FieldNameCaseCamel) + ":",
			fmt.Sprintf(` #"%s",`, field.Name),
		}
	}
	tw := tablewriter.NewWriter(buffer)
	tw.SetBorder(false)
	tw.SetRowLine(false)
	tw.SetAutoWrapText(false)
	tw.SetColumnSeparator("")
	tw.AppendBulk(array)
	tw.Render()
	namesContent := buffer.String()
	// Let's do this hack of table writer for indent!
	namesContent = gstr.Replace(namesContent, "  #", "")
	buffer.Reset()
	buffer.WriteString(namesContent)
	return buffer.String()
}

// generateColumnDefinitionForDao generates and returns the column names definition for specified table.
func generateColumnDefinitionForDao(fieldMap map[string]*gdb.TableField, removeFieldPrefixArray []string) string {
	var (
		buffer = bytes.NewBuffer(nil)
		array  = make([][]string, len(fieldMap))
		names  = sortFieldKeyForDao(fieldMap)
	)

	for index, name := range names {
		var (
			field   = fieldMap[name]
			comment = gstr.Trim(gstr.ReplaceByArray(field.Comment, g.SliceStr{
				"\n", " ",
				"\r", " ",
			}))
		)
		newFiledName := field.Name
		for _, v := range removeFieldPrefixArray {
			newFiledName = gstr.TrimLeftStr(newFiledName, v, 1)
		}
		array[index] = []string{
			"    #" + formatFieldName(newFiledName, FieldNameCaseCamel),
			" # " + "string",
			" #" + fmt.Sprintf(`// %s`, comment),
		}
	}
	tw := tablewriter.NewWriter(buffer)
	tw.SetBorder(false)
	tw.SetRowLine(false)
	tw.SetAutoWrapText(false)
	tw.SetColumnSeparator("")
	tw.AppendBulk(array)
	tw.Render()
	defineContent := buffer.String()
	// Let's do this hack of table writer for indent!
	defineContent = gstr.Replace(defineContent, "  #", "")
	buffer.Reset()
	buffer.WriteString(defineContent)
	return buffer.String()
}
