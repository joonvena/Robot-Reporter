package main

import (
	"encoding/xml"
)

type Output struct {
	XMLName   xml.Name `xml:"robot"`
	Text      string   `xml:",chardata"`
	Generator string   `xml:"generator,attr"`
	Generated string   `xml:"generated,attr"`
	Rpa       string   `xml:"rpa,attr"`
	Suite     struct {
		Text   string `xml:",chardata"`
		ID     string `xml:"id,attr"`
		Name   string `xml:"name,attr"`
		Source string `xml:"source,attr"`
		Suite  struct {
			Text   string `xml:",chardata"`
			ID     string `xml:"id,attr"`
			Name   string `xml:"name,attr"`
			Source string `xml:"source,attr"`
			Test   []struct {
				Text string `xml:",chardata"`
				ID   string `xml:"id,attr"`
				Name string `xml:"name,attr"`
				Kw   []struct {
					Text      string `xml:",chardata"`
					Name      string `xml:"name,attr"`
					Library   string `xml:"library,attr"`
					Type      string `xml:"type,attr"`
					Doc       string `xml:"doc"`
					Arguments struct {
						Text string   `xml:",chardata"`
						Arg  []string `xml:"arg"`
					} `xml:"arguments"`
					Msg struct {
						Text      string `xml:",chardata"`
						Timestamp string `xml:"timestamp,attr"`
						Level     string `xml:"level,attr"`
					} `xml:"msg"`
					Status struct {
						Text      string `xml:",chardata"`
						Status    string `xml:"status,attr"`
						Starttime string `xml:"starttime,attr"`
						Endtime   string `xml:"endtime,attr"`
					} `xml:"status"`
				} `xml:"kw"`
				Status struct {
					Text      string `xml:",chardata"`
					Status    string `xml:"status,attr"`
					Starttime string `xml:"starttime,attr"`
					Endtime   string `xml:"endtime,attr"`
					Critical  string `xml:"critical,attr"`
				} `xml:"status"`
			} `xml:"test"`
			Status struct {
				Text      string `xml:",chardata"`
				Status    string `xml:"status,attr"`
				Starttime string `xml:"starttime,attr"`
				Endtime   string `xml:"endtime,attr"`
			} `xml:"status"`
		} `xml:"suite"`
		Status struct {
			Text      string `xml:",chardata"`
			Status    string `xml:"status,attr"`
			Starttime string `xml:"starttime,attr"`
			Endtime   string `xml:"endtime,attr"`
		} `xml:"status"`
	} `xml:"suite"`
	Statistics struct {
		Text  string `xml:",chardata"`
		Total struct {
			Text string `xml:",chardata"`
			Stat []struct {
				Text string `xml:",chardata"`
				Pass string `xml:"pass,attr"`
				Fail string `xml:"fail,attr"`
				Skip string `xml:"skip,attr"`
			} `xml:"stat"`
		} `xml:"total"`
		Tag   string `xml:"tag"`
		Suite struct {
			Text string `xml:",chardata"`
			Stat []struct {
				Text string `xml:",chardata"`
				Pass string `xml:"pass,attr"`
				Fail string `xml:"fail,attr"`
				ID   string `xml:"id,attr"`
				Name string `xml:"name,attr"`
			} `xml:"stat"`
		} `xml:"suite"`
	} `xml:"statistics"`
	Errors string `xml:"errors"`
}

type Test struct {
	Name          string `json:"name"`
	Message       string `json:"message"`
	ExecutionTime string `json:"executionTime"`
	Suite         string `json:"suite"`
}
