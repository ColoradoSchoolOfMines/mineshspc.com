package database

func (d *Database) GetAllEmails(ctx context.Context) ([]string, error) {
	row := d.DB.Query(ctx, `
		SELECT teacheremail, t.name
		FROM teams t
		UNION
		SELECT email, namw
		FROM students s
	`, email, teamID)
	return d.scanTeamWithStudents(ctx, row)
}