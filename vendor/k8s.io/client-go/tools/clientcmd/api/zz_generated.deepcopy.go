// +build !ignore_autogenerated

/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This file was autogenerated by deepcopy-gen. Do not edit it manually!

package api

import (
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
	reflect "reflect"
)

// Deprecated: register deep-copy functions.
func init() {
	SchemeBuilder.Register(RegisterDeepCopies)
}

// Deprecated: RegisterDeepCopies adds deep-copy functions to the given scheme. Public
// to allow building arbitrary schemes.
func RegisterDeepCopies(scheme *runtime.Scheme) error {
	return scheme.AddGeneratedDeepCopyFuncs(
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*AuthInfo).DeepCopyInto(out.(*AuthInfo))
			return nil
		}, InType: reflect.TypeOf(&AuthInfo{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*AuthProviderConfig).DeepCopyInto(out.(*AuthProviderConfig))
			return nil
		}, InType: reflect.TypeOf(&AuthProviderConfig{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*Cluster).DeepCopyInto(out.(*Cluster))
			return nil
		}, InType: reflect.TypeOf(&Cluster{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*Config).DeepCopyInto(out.(*Config))
			return nil
		}, InType: reflect.TypeOf(&Config{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*Context).DeepCopyInto(out.(*Context))
			return nil
		}, InType: reflect.TypeOf(&Context{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*Preferences).DeepCopyInto(out.(*Preferences))
			return nil
		}, InType: reflect.TypeOf(&Preferences{})},
	)
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuthInfo) DeepCopyInto(out *AuthInfo) {
	*out = *in
	if in.ClientCertificateData != nil {
		in, out := &in.ClientCertificateData, &out.ClientCertificateData
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.ClientKeyData != nil {
		in, out := &in.ClientKeyData, &out.ClientKeyData
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.ImpersonateGroups != nil {
		in, out := &in.ImpersonateGroups, &out.ImpersonateGroups
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ImpersonateUserExtra != nil {
		in, out := &in.ImpersonateUserExtra, &out.ImpersonateUserExtra
		*out = make(map[string][]string, len(*in))
		for key, val := range *in {
			if val == nil {
				(*out)[key] = nil
			} else {
				(*out)[key] = make([]string, len(val))
				copy((*out)[key], val)
			}
		}
	}
	if in.AuthProvider != nil {
		in, out := &in.AuthProvider, &out.AuthProvider
		if *in == nil {
			*out = nil
		} else {
			*out = new(AuthProviderConfig)
			(*in).DeepCopyInto(*out)
		}
	}
	if in.Extensions != nil {
		in, out := &in.Extensions, &out.Extensions
		*out = make(map[string]runtime.Object, len(*in))
		for key, val := range *in {
			if val == nil {
				(*out)[key] = nil
			} else {
				(*out)[key] = val.DeepCopyObject()
			}
		}
	}
	return
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, creating a new AuthInfo.
func (x *AuthInfo) DeepCopy() *AuthInfo {
	if x == nil {
		return nil
	}
	out := new(AuthInfo)
	x.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuthProviderConfig) DeepCopyInto(out *AuthProviderConfig) {
	*out = *in
	if in.Config != nil {
		in, out := &in.Config, &out.Config
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, creating a new AuthProviderConfig.
func (x *AuthProviderConfig) DeepCopy() *AuthProviderConfig {
	if x == nil {
		return nil
	}
	out := new(AuthProviderConfig)
	x.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Cluster) DeepCopyInto(out *Cluster) {
	*out = *in
	if in.CertificateAuthorityData != nil {
		in, out := &in.CertificateAuthorityData, &out.CertificateAuthorityData
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.Extensions != nil {
		in, out := &in.Extensions, &out.Extensions
		*out = make(map[string]runtime.Object, len(*in))
		for key, val := range *in {
			if val == nil {
				(*out)[key] = nil
			} else {
				(*out)[key] = val.DeepCopyObject()
			}
		}
	}
	return
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, creating a new Cluster.
func (x *Cluster) DeepCopy() *Cluster {
	if x == nil {
		return nil
	}
	out := new(Cluster)
	x.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Config) DeepCopyInto(out *Config) {
	*out = *in
	in.Preferences.DeepCopyInto(&out.Preferences)
	if in.Clusters != nil {
		in, out := &in.Clusters, &out.Clusters
		*out = make(map[string]*Cluster, len(*in))
		for key, val := range *in {
			if val == nil {
				(*out)[key] = nil
			} else {
				(*out)[key] = new(Cluster)
				val.DeepCopyInto((*out)[key])
			}
		}
	}
	if in.AuthInfos != nil {
		in, out := &in.AuthInfos, &out.AuthInfos
		*out = make(map[string]*AuthInfo, len(*in))
		for key, val := range *in {
			if val == nil {
				(*out)[key] = nil
			} else {
				(*out)[key] = new(AuthInfo)
				val.DeepCopyInto((*out)[key])
			}
		}
	}
	if in.Contexts != nil {
		in, out := &in.Contexts, &out.Contexts
		*out = make(map[string]*Context, len(*in))
		for key, val := range *in {
			if val == nil {
				(*out)[key] = nil
			} else {
				(*out)[key] = new(Context)
				val.DeepCopyInto((*out)[key])
			}
		}
	}
	if in.Extensions != nil {
		in, out := &in.Extensions, &out.Extensions
		*out = make(map[string]runtime.Object, len(*in))
		for key, val := range *in {
			if val == nil {
				(*out)[key] = nil
			} else {
				(*out)[key] = val.DeepCopyObject()
			}
		}
	}
	return
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, creating a new Config.
func (x *Config) DeepCopy() *Config {
	if x == nil {
		return nil
	}
	out := new(Config)
	x.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (x *Config) DeepCopyObject() runtime.Object {
	if c := x.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Context) DeepCopyInto(out *Context) {
	*out = *in
	if in.Extensions != nil {
		in, out := &in.Extensions, &out.Extensions
		*out = make(map[string]runtime.Object, len(*in))
		for key, val := range *in {
			if val == nil {
				(*out)[key] = nil
			} else {
				(*out)[key] = val.DeepCopyObject()
			}
		}
	}
	return
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, creating a new Context.
func (x *Context) DeepCopy() *Context {
	if x == nil {
		return nil
	}
	out := new(Context)
	x.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Preferences) DeepCopyInto(out *Preferences) {
	*out = *in
	if in.Extensions != nil {
		in, out := &in.Extensions, &out.Extensions
		*out = make(map[string]runtime.Object, len(*in))
		for key, val := range *in {
			if val == nil {
				(*out)[key] = nil
			} else {
				(*out)[key] = val.DeepCopyObject()
			}
		}
	}
	return
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, creating a new Preferences.
func (x *Preferences) DeepCopy() *Preferences {
	if x == nil {
		return nil
	}
	out := new(Preferences)
	x.DeepCopyInto(out)
	return out
}
