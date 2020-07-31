package parse

import (
	"bufio"
	"encoding/json"
	"time"

	"pressure-test/base"
)

func GetTransaction() *base.TransactionInstance {
	r := map[string]interface{}{
		"id":       0,
		"name":     "apm",
		"password": "123456",
	}
	rBytes, _ := json.Marshal(r)
	return &base.TransactionInstance{
		Data: base.Data{
			Id:                 "",
			Name:               "",
			PresetTransactions: []*base.PresetCase{},
			DoTransactions: []*base.Case{
				{
					Clients: []base.Client{
						&base.FastHttp{
							Url:    "http://10.78.74.37:8099/test", //192.168.86.113;10.78.74.37
							Method: "POST",
							Header: map[string]string{
								"Content-Type": "application/json;charset=UTF-8",
							},
							TimeOut: 10 * time.Second,
							Body: func(w *bufio.Writer) {
								_, _ = w.Write(rBytes)
							},
						},
					},
					RunMode: &base.DurationMode{GradientModes: []*base.GradientMode{
						{
							Num:          15000,
							DurationTime: 30 * time.Second,
						},
					}},
				},
			},
		},
	}
}
