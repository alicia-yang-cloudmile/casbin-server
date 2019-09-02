// Copyright 2018 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"errors"
	"log"

	pb "github.com/casbin/casbin-server/proto"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

// Server is used to implement proto.CasbinServer.
type Server struct {
	enforcerMap map[string]*casbin.Enforcer
	adapterMap  map[string]persist.Adapter
}

func NewServer() *Server {
	s := Server{}

	s.enforcerMap = map[string]*casbin.Enforcer{}
	s.adapterMap = map[string]persist.Adapter{}

	return &s
}

func (s *Server) getEnforcer(handle string) (*casbin.Enforcer, error) {
	if _, ok := s.enforcerMap[handle]; ok {
		return s.enforcerMap[handle], nil
	} else {
		return nil, errors.New("enforcer not found")
	}
}

func (s *Server) getAdapter(handle string) (persist.Adapter, error) {
	if _, ok := s.adapterMap[handle]; ok {
		return s.adapterMap[handle], nil
	} else {
		return nil, errors.New("adapter not found")
	}
}

func (s *Server) addEnforcer(e *casbin.Enforcer, handle string){
	s.enforcerMap[handle] = e
}

func (s *Server) addAdapter(a persist.Adapter, handle string) {
	s.adapterMap[handle] = a
}

func (s *Server) NewEnforcer(ctx context.Context, in *pb.NewEnforcerRequest) (*pb.NewEnforcerReply, error) {
	var a persist.Adapter
	var m model.Model
	var e *casbin.Enforcer
	var err error

	handler := in.EnforcerName
	if handler == "" {
		return nil, errors.New("Bad NewEnforcerRequest")
	}
	_, err = s.getEnforcer(handler)
	if err != nil {
		log.Println("Create New Enforcer")
		a, err = s.getAdapter(in.AdapterHandle)
		if err != nil {
			return &pb.NewEnforcerReply{Handler: ""}, err
		}
		m, err = model.NewModelFromString(in.ModelText)
		if err != nil {
			return &pb.NewEnforcerReply{Handler: ""}, err
		}
		e, err = casbin.NewEnforcer(m, a)
		if err != nil {
			return &pb.NewEnforcerReply{Handler: ""}, err
		}
		s.addEnforcer(e, handler)
	}
	return &pb.NewEnforcerReply{Handler: handler}, nil
	
}

func (s *Server) NewAdapter(ctx context.Context, in *pb.NewAdapterRequest) (*pb.NewAdapterReply, error) {
	var a persist.Adapter
	var err error

	handler := in.AdapterName
	if handler == "" {
		return nil, errors.New("Bad NewAdapterRequest")
	}
	_, err = s.getAdapter(handler)
	if err != nil {
		log.Println("Create New Adapter")
		a, err = newAdapter(in)
		if err != nil {
			return nil, err
		}
		s.addAdapter(a, handler)
	}
	return &pb.NewAdapterReply{Handler: handler}, nil
}

func (s *Server) Enforce(ctx context.Context, in *pb.EnforceRequest) (*pb.BoolReply, error) {
	e, err := s.getEnforcer(in.EnforcerHandler)
	if err != nil {
		return &pb.BoolReply{Res: false}, err
	}

	params := make([]interface{}, 0, len(in.Params))

	m := e.GetModel()["m"]["m"]
	sourceValue := m.Value

	for index := range in.Params {
		param := parseAbacParam(in.Params[index], m)
		params = append(params, param)
	}

	res, err := e.Enforce(params...)
	if err != nil {
		return &pb.BoolReply{Res: false}, err
	}

	m.Value = sourceValue

	return &pb.BoolReply{Res: res}, nil
}

func (s *Server) LoadPolicy(ctx context.Context, in *pb.EmptyRequest) (*pb.EmptyReply, error) {
	e, err := s.getEnforcer(in.Handler)
	if err != nil {
		return &pb.EmptyReply{}, err
	}

	err = e.LoadPolicy()

	return &pb.EmptyReply{}, err
}

func (s *Server) SavePolicy(ctx context.Context, in *pb.EmptyRequest) (*pb.EmptyReply, error) {
	e, err := s.getEnforcer(in.Handler)
	if err != nil {
		return &pb.EmptyReply{}, err
	}

	err = e.SavePolicy()

	return &pb.EmptyReply{}, err
}
