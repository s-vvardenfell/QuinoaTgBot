package client

import (
	"fmt"

	"github.com/s-vvardenfell/QuinoaTgBot/conditions"
)

//клиент используется для взаимодействия с сервером
//посредством REST API
type Client struct {
}

func New() *Client {
	return &Client{}
}

func (c *Client) FilmsByConditions(cnd conditions.Conditions) string {
	return fmt.Sprintf("PARSED FILMS BY YOUR CONDITIONS: %s", cnd.Type)
}
