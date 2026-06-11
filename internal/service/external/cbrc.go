package external

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/beevik/etree"
)

type CBRClient struct {
	client *http.Client
	xmlURL string
}

func NewCBRClient() *CBRClient {
	return &CBRClient{
		client: &http.Client{Timeout: 10 * time.Second},
		xmlURL: "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx",
	}
}

func (c *CBRClient) GetKeyRate(ctx context.Context) (float64, error) {
	// SOAP конверт для метода GetKeyRate
	soapEnvelope := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <GetKeyRate xmlns="http://web.cbr.ru/" />
  </soap:Body>
</soap:Envelope>`

	req, err := http.NewRequestWithContext(ctx, "POST", c.xmlURL, bytes.NewBufferString(soapEnvelope))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://web.cbr.ru/GetKeyRate")

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// Парсинг XML через beevik/etree
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(body); err != nil {
		return 0, err
	}

	// Поиск элемента с ключевой ставкой (в структуре ответа ЦБ РФ)
	// Элемент обычно называется KeyRate внутри Диффузных таблиц/ответа
	rateElem := doc.FindElement("//KeyRate")
	if rateElem == nil {
		// Альтернативный поиск по локальному имени узла в схемах SOAP
		rateElem = doc.FindElement("//MainRate")
		if rateElem == nil {
			return 0, fmt.Errorf("failed to find KeyRate element in XML response")
		}
	}

	rate, err := strconv.ParseFloat(rateElem.Text(), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse rate text: %w", err)
	}

	return rate, nil
}
