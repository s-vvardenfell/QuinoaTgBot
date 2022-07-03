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

var (
	ErrNoResults   = errors.New("по вашему запросу не найдено подходящих результатов")
	ErrSomeError   = errors.New("во время обработки запроса произошла непредвиденная ошибка")
	contextTimeout = 600
)

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(contextTimeout)*time.Second)
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
		if strings.Contains(err.Error(), "no results found") ||
			strings.Contains(err.Error(), "no conditions") {
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
