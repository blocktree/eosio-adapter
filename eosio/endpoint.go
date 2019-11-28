/*
 * Copyright 2019 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package eosio

import (
	"fmt"
	"github.com/blocktree/openwallet/log"
	"github.com/imroc/req"
)

type Client struct {
	BaseURL string
	Debug   bool
	client  *req.Req
}

func NewClient(url string, debug bool) *Client {
	c := Client{
		BaseURL: url,
		Debug:   debug,
	}

	api := req.New()
	c.client = api
	return &c
}

//SendRPCRequest 发起JSON-RPC请求
//@optional
func (c *Client) SendRPCRequest(method string, request interface{}) ([]byte, error) {

	if c.client == nil {
		return nil, fmt.Errorf("API url is not setup. ")
	}

	if c.Debug {
		log.Std.Info("Start Request API...")
	}
	targetURL := fmt.Sprintf("%s/v1/chain/%s", c.BaseURL, method)
	r, err := c.client.Post(targetURL, req.BodyJSON(&request))

	if c.Debug {
		log.Std.Info("Request API Completed")
	}

	if c.Debug {
		log.Std.Info("%+v", r)
	}

	if err != nil {
		return nil, err
	}
	return r.Bytes(), nil
}

//SupportJsonRPCEndpoint 是否开放客户端直接调用全节点的JSON-RPC方法
//@optional
func (c *Client) SupportJsonRPCEndpoint() bool {
	return true
}
