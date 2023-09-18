package controllers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/eonebyte/myapp/app/config"
	"github.com/eonebyte/myapp/app/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type UserController struct {
	baseUrl string
	session *session.Store
}

func NewUserController(session *session.Store, b string) *UserController {
	return &UserController{
		session: session,
		baseUrl: b,
	}
}

func (u *UserController) PageUser(c *fiber.Ctx) error {
	session, err := u.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}

	//get Session Data before Decode
	sessionDataJSON := session.Get("sessionData")

	// Convert the JSON data back to the User struct
	var sessionData SessionData
	err = json.Unmarshal(sessionDataJSON.([]byte), &sessionData)
	if err != nil {
		return c.SendString("Failed to unmarshal session data")
	}
	sessionData.UrlPath = u.baseUrl + "/users"
	var getUsers models.User
	users, err := getUsers.Gets()
	if err != nil {
		return err
	}
	data := fiber.Map{
		"sessionData": sessionData,
		"users":       users,
	}
	return c.Render("user", data, "layouts/admin")
}

func (u *UserController) ProfilUser(c *fiber.Ctx) error {
	session, err := u.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}
	sessionDataJSON := session.Get("sessionData")

	// Convert the JSON data back to the User struct
	var sessionData SessionData
	err = json.Unmarshal(sessionDataJSON.([]byte), &sessionData)
	if err != nil {
		return c.SendString("Failed to unmarshal session data")
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", sessionData.UserData.Id).Error; err != nil {
		return err
	}

	data := fiber.Map{
		"sessionData": sessionData,
		"User":        user,
	}

	return c.Render("profil_user", data, "layouts/admin")
}

func (u *UserController) UpdateProfilUser(c *fiber.Ctx) error {
	reqUser := new(models.User)
	err := c.BodyParser(reqUser)
	if err != nil {
		return err
	}
	user, err := reqUser.Get(reqUser.Id)
	if err != nil {
		return err
	}

	if reqUser.Name != "" {
		user.Name = reqUser.Name
	}
	if reqUser.Email != "" {
		user.Email = reqUser.Email
	}
	if reqUser.Phone != "" {
		user.Phone = reqUser.Phone
	}
	if reqUser.Address != "" {
		user.Address = reqUser.Address
	}
	file, err := c.FormFile("image")
	if file != nil {
		//revome old file image
		oldImageName := user.Image
		if oldImageName != "" {
			err = os.Remove(filepath.Join("static/img/users/", oldImageName))
			if err != nil {
				return err
			}
		}

		// Save file to root directory images:
		fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))
		c.SaveFile(file, fmt.Sprintf("./static/img/users/%s", fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))))
		if err != nil {
			return err
		}
		user.Image = fileName
	}
	if reqUser.Password != "" {
		hashedPassword := hashPassword(reqUser.Password)
		user.Password = string(hashedPassword)
	}
	user.Update()

	

	return c.Redirect("/dashboard/profil-user")
}

func (u *UserController) UpdateUser(c *fiber.Ctx) error {
	session, err := u.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}
	reqUser := new(models.User)
	err = c.BodyParser(reqUser)
	if err != nil {
		return err
	}
	user, err := reqUser.Get(reqUser.Id)
	if err != nil {
		return err
	}
	if reqUser.Name != "" {
		user.Name = reqUser.Name
	}
	if reqUser.Email != "" {
		user.Email = reqUser.Email
	}
	if reqUser.Phone != "" {
		user.Phone = reqUser.Phone
	}
	user.UpdatedAt = time.Now()
	user.Update()

	return c.Redirect("/dashboard/users")
}

func (u *UserController) DeleteUser(c *fiber.Ctx) error {
	session, err := u.session.Get(c)
	if err != nil {
		return err
	}
	if session.Get("authenticated") != true {
		return c.Redirect("/auth")
	}
	reqUserDelete := new(models.User)
	err = c.BodyParser(reqUserDelete)
	if err != nil {
		return err
	}
	user, err := reqUserDelete.Get(reqUserDelete.Id)
	if err != nil {
		return err
	}
	user.Delete()

	return c.Redirect("/dashboard/users")
}
