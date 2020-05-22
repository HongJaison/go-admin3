package datamodel

import (
	"github.com/HongJaison/go-admin3/context"
	"github.com/HongJaison/go-admin3/modules/db"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/table"
	"github.com/HongJaison/go-admin3/template"
	"github.com/HongJaison/go-admin3/template/types"
	"github.com/HongJaison/go-admin3/template/types/form"
	editType "github.com/HongJaison/go-admin3/template/types/table"
	template2 "html/template"
)

// GetPostsTable return the model of table posts.
func GetPostsTable(ctx *context.Context) (postsTable table.Table) {

	postsTable = table.NewDefaultTable(table.DefaultConfig().SetExportable(true))

	info := postsTable.GetInfo()
	info.AddField("ID", "id", db.Int).FieldSortable()
	info.AddField("Title", "title", db.Varchar)
	info.AddField("AuthorID", "author_id", db.Varchar).FieldDisplay(func(value types.FieldModel) interface{} {
		return template.Default().
			Link().
			SetURL("/admin/info/authors/detail?__goadmin_detail_pk=" + value.Value).
			SetContent(template2.HTML(value.Value)).
			OpenInNewTab().
			SetTabTitle("Author Detail").
			GetContent()
	})
	info.AddField("AuthorName", "name", db.Varchar).FieldDisplay(func(value types.FieldModel) interface{} {
		first, _ := value.Row["authors_goadmin_join_first_name"].(string)
		last, _ := value.Row["authors_goadmin_join_last_name"].(string)
		return first + " " + last
	})
	info.AddField("AuthorFirstName", "first_name", db.Varchar).FieldJoin(types.Join{
		Field:     "author_id",
		JoinField: "id",
		Table:     "authors",
	}).FieldHide()
	info.AddField("AuthorLastName", "last_name", db.Varchar).FieldJoin(types.Join{
		Field:     "author_id",
		JoinField: "id",
		Table:     "authors",
	}).FieldHide()
	info.AddField("Description", "description", db.Varchar)
	info.AddField("Content", "content", db.Varchar).FieldEditAble(editType.Textarea)
	info.AddField("Date", "date", db.Varchar)
	info.AddField("AuthorCreatedAt", "created_at", db.Varchar).FieldJoin(types.Join{
		Field:     "author_id",
		JoinField: "id",
		Table:     "authors",
	})
	info.AddField("Created_at", "created_at", db.Varchar).FieldFilterable()

	info.SetTable("posts").SetTitle("Posts").SetDescription("Posts")

	formList := postsTable.GetForm()
	formList.AddField("ID", "id", db.Int, form.Default).FieldNotAllowEdit().FieldNotAllowAdd()
	formList.AddField("Title", "title", db.Varchar, form.Text)
	formList.AddField("Description", "description", db.Varchar, form.Text)
	formList.AddField("Content", "content", db.Varchar, form.RichText).FieldEnableFileUpload()
	formList.AddField("Date", "date", db.Varchar, form.Datetime)
	formList.EnableAjax("Request Success", "Request Failed")

	formList.SetTable("posts").SetTitle("Posts").SetDescription("Posts")

	return
}
