package admin

import (
	"opensavecloudserver/database"
	"opensavecloudserver/upload"
)

// RemoveUser rome the user from the db and all his datas
func RemoveUser(user *database.User) error {
	if err := database.RemoveAllUserGameEntries(user); err != nil {
		return err
	}
	if err := upload.RemoveFolders(user.ID); err != nil {
		return err
	}
	return database.RemoveUser(user)
}

func SetAdmin(user *database.User) error {
	user.Role = database.AdminRole
	user.IsAdmin = true
	return database.SaveUser(user)
}

func RemoveAdminRole(user *database.User) error {
	user.Role = database.UserRole
	user.IsAdmin = false
	return database.SaveUser(user)
}
