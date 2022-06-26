package client

import (
	"context"
	"fmt"
	"time"

	"github.com/s-vvardenfell/QuinoaTgBot/conditions"
	"github.com/s-vvardenfell/QuinoaTgBot/generated"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type QuinoaTgBotClient struct {
	generated.MainServiceClient
}

func New(host, port string) *QuinoaTgBotClient {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", host, port), grpc.WithInsecure())
	if err != nil {
		logrus.Fatalf("cannot connect to host< %s> and port <%s>: %v", host, port, err)
	}
	return &QuinoaTgBotClient{
		generated.NewMainServiceClient(conn),
	}
}

func (c *QuinoaTgBotClient) FilmsByConditions(cnd conditions.Conditions) string {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resFromServ, err := c.GetParsedData(ctx, &generated.Conditions{
		Type:      cnd.Type,
		Genres:    cnd.Genres,
		StartYear: cnd.StartYear,
		EndYear:   cnd.EndYear,
		Countries: cnd.Countries,
		Keyword:   cnd.Keyword,
	})

	if err != nil {
		return "Во время выполнения запроса произошла ошибка"
	}
	return resFromServ.String()
}
