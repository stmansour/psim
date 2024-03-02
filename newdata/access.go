package newdata

import (
	"fmt"

	// Register the MySQL driver
	_ "github.com/go-sql-driver/mysql"
)

// GrantReadAccess grants read access to all tables in the specified database
// for a list of usernames.
// -----------------------------------------------------------------------------
func (p *DatabaseSQL) GrantReadAccess(usernames []string) error {
	for _, username := range usernames {
		grantStmt := fmt.Sprintf("GRANT SELECT ON %s.* TO '%s'@'localhost'", p.Name, username)
		if _, err := p.DB.Exec(grantStmt); err != nil {
			return fmt.Errorf("failed to grant read access to %s: %v", username, err)
		}
	}
	if _, err := p.DB.Exec("FLUSH PRIVILEGES"); err != nil {
		return fmt.Errorf("failed to flush privileges: %v", err)
	}
	return nil
}

// GrantFullAccess grants full access to the 'plato' database for a list of
// usernames.
// --------------------------------------------------------------------------
func (p *DatabaseSQL) GrantFullAccess(usernames []string) error {
	for _, username := range usernames {
		grantStmt := fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO '%s'@'localhost'", p.Name, username)
		if _, err := p.DB.Exec(grantStmt); err != nil {
			return fmt.Errorf("failed to grant privileges to %s: %v", username, err)
		}
	}
	if _, err := p.DB.Exec("FLUSH PRIVILEGES"); err != nil {
		return fmt.Errorf("failed to flush privileges: %v", err)
	}
	return nil
}
