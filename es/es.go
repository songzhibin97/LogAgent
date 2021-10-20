package es

import "github.com/olivere/elastic/v7"

func InitEs(address string) (client *elastic.Client, err error) {
	return elastic.NewClient(elastic.SetURL(address))
}
