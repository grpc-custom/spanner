package ddl

func GenerateDatabaseCode(table *Table) {

}

var (
	createDatabaseTemplate = `
func Create{{.DatabaseName}}Database(ctx context.Context, client *database.DatabaseAdminClient, projectID, instance, database string) error {
	op, err := client.CreateDatabase(ctx, &adminpb.CreateDatabaseRequest{
		Parent: fmt.Sprintf("projects/%s/instances/%s", projectID, instance),
		CreateStatement: fmt.Sprintf("CREATE DATABASE %s", database),
		ExtraStatements: []string{},
	})
	if err != nil {
		return err
	}
	if _, err := op.Wait(ctx); err != nil {
		return err
	}
	return nil
}
`
)
