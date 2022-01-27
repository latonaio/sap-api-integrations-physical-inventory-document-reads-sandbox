package sap_api_caller

import (
	"fmt"
	"io/ioutil"
	"net/http"
	sap_api_output_formatter "sap-api-integrations-physical-inventory-document-reads/SAP_API_Output_Formatter"
	"strings"
	"sync"

	"github.com/latonaio/golang-logging-library-for-sap/logger"
	"golang.org/x/xerrors"
)

type SAPAPICaller struct {
	baseURL string
	apiKey  string
	log     *logger.Logger
}

func NewSAPAPICaller(baseUrl string, l *logger.Logger) *SAPAPICaller {
	return &SAPAPICaller{
		baseURL: baseUrl,
		apiKey:  GetApiKey(),
		log:     l,
	}
}

func (c *SAPAPICaller) AsyncGetPhysicalInventoryDocument(fiscalYear, physicalInventoryDocument, physicalInventoryDocumentItem string, accepter []string) {
	wg := &sync.WaitGroup{}
	wg.Add(len(accepter))
	for _, fn := range accepter {
		switch fn {
		case "Header":
			func() {
				c.Header(fiscalYear, physicalInventoryDocument)
				wg.Done()
			}()
		case "Item":
			func() {
				c.Item(fiscalYear, physicalInventoryDocument, physicalInventoryDocumentItem)
				wg.Done()
			}()
		default:
			wg.Done()
		}
	}

	wg.Wait()
}

func (c *SAPAPICaller) Header(fiscalYear, physicalInventoryDocument string) {
	headerData, err := c.callPhysicalInventoryDocumentSrvAPIRequirementHeader("A_PhysInventoryDocHeader", fiscalYear, physicalInventoryDocument)
	if err != nil {
		c.log.Error(err)
		return
	}
	c.log.Info(headerData)

	itemData, err := c.callToItem(headerData[0].ToItem)
	if err != nil {
		c.log.Error(err)
		return
	}
	c.log.Info(itemData)

}

func (c *SAPAPICaller) callPhysicalInventoryDocumentSrvAPIRequirementHeader(api, fiscalYear, physicalInventoryDocument string) ([]sap_api_output_formatter.Header, error) {
	url := strings.Join([]string{c.baseURL, "API_PHYSICAL_INVENTORY_DOC_SRV", api}, "/")
	req, _ := http.NewRequest("GET", url, nil)

	c.setHeaderAPIKeyAccept(req)
	c.getQueryWithHeader(req, fiscalYear, physicalInventoryDocument)

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return nil, xerrors.Errorf("API request error: %w", err)
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	data, err := sap_api_output_formatter.ConvertToHeader(byteArray, c.log)
	if err != nil {
		return nil, xerrors.Errorf("convert error: %w", err)
	}
	return data, nil
}

func (c *SAPAPICaller) callToItem(url string) ([]sap_api_output_formatter.ToItem, error) {
	req, _ := http.NewRequest("GET", url, nil)
	c.setHeaderAPIKeyAccept(req)

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return nil, xerrors.Errorf("API request error: %w", err)
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	data, err := sap_api_output_formatter.ConvertToToItem(byteArray, c.log)
	if err != nil {
		return nil, xerrors.Errorf("convert error: %w", err)
	}
	return data, nil
}

func (c *SAPAPICaller) Item(fiscalYear, physicalInventoryDocument, physicalInventoryDocumentItem string) {
	itemData, err := c.callPhysicalInventoryDocumentSrvAPIRequirementItem("A_PhysInventoryDocItem", fiscalYear, physicalInventoryDocument, physicalInventoryDocumentItem)
	if err != nil {
		c.log.Error(err)
		return
	}
	c.log.Info(itemData)

}

func (c *SAPAPICaller) callPhysicalInventoryDocumentSrvAPIRequirementItem(api, fiscalYear, physicalInventoryDocument, physicalInventoryDocumentItem string) ([]sap_api_output_formatter.Item, error) {
	url := strings.Join([]string{c.baseURL, "API_PHYSICAL_INVENTORY_DOC_SRV", api}, "/")
	req, _ := http.NewRequest("GET", url, nil)

	c.setHeaderAPIKeyAccept(req)
	c.getQueryWithItem(req, fiscalYear, physicalInventoryDocument, physicalInventoryDocumentItem)

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return nil, xerrors.Errorf("API request error: %w", err)
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	data, err := sap_api_output_formatter.ConvertToItem(byteArray, c.log)
	if err != nil {
		return nil, xerrors.Errorf("convert error: %w", err)
	}
	return data, nil
}

func (c *SAPAPICaller) setHeaderAPIKeyAccept(req *http.Request) {
	req.Header.Set("APIKey", c.apiKey)
	req.Header.Set("Accept", "application/json")
}

func (c *SAPAPICaller) getQueryWithHeader(req *http.Request, fiscalYear, physicalInventoryDocument string) {
	params := req.URL.Query()
	params.Add("$filter", fmt.Sprintf("FiscalYear eq '%s' and PhysicalInventoryDocument eq '%s'", fiscalYear, physicalInventoryDocument))
	req.URL.RawQuery = params.Encode()
}

func (c *SAPAPICaller) getQueryWithItem(req *http.Request, fiscalYear, physicalInventoryDocument, physicalInventoryDocumentItem string) {
	params := req.URL.Query()
	params.Add("$filter", fmt.Sprintf("FiscalYear eq '%s' and PhysicalInventoryDocument eq '%s' and PhysicalInventoryDocumentItem eq '%s'", fiscalYear, physicalInventoryDocument, physicalInventoryDocumentItem))
	req.URL.RawQuery = params.Encode()
}
