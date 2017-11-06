package controllers

import (
	"github.com/revel/revel"
)

type Application struct {
	*revel.Controller
}

func (c Application) Index() revel.Result {
	return c.Render()
}

func (c Application) EnterDemo(room int, user string) revel.Result {
	c.Validation.Required(room)
	c.Validation.Required(user)

	if c.Validation.HasErrors() {
		c.Flash.Error("Please choose a nick name and the demonstration type.")
		return c.Redirect(Application.Index)
	}

	return c.Redirect("/websocket/room?room=%s&user=%s", room, user)
}
