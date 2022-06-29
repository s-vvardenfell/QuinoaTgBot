package client

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/s-vvardenfell/QuinoaTgBot/conditions"
	"github.com/s-vvardenfell/QuinoaTgBot/generated"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var ErrNoResults = errors.New("по вашему запросу не найдено подходящих результатов")
var ErrSomeError = errors.New("во время обработки запроса произошла непредвиденная ошибка")

type ParsingResults struct {
	Name string
	Ref  string
	Img  string
}

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

func (c *QuinoaTgBotClient) FilmsByConditions(
	cnd conditions.Conditions) ([]ParsingResults, error) {
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
		if strings.Contains(err.Error(), "no results found") {
			return nil, ErrNoResults
		}
		return nil, ErrSomeError
	}

	res := make([]ParsingResults, 0, len(resFromServ.Data))

	for i := range resFromServ.Data {
		res = append(res, ParsingResults{
			Name: resFromServ.Data[i].Name,
			Ref:  resFromServ.Data[i].Ref,
			Img:  resFromServ.Data[i].Img,
		})
	}

	return res, nil
}
