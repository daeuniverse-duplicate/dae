/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2022-2023, v2rayA Organization <team@v2raya.org>
 */

package netutils

import (
	"context"
	"errors"
	"fmt"
	"github.com/mzz2017/softwind/netproxy"
	"golang.org/x/net/dns/dnsmessage"
	"net/netip"
	"sync"
)

type Ip46 struct {
	Ip4 netip.Addr
	Ip6 netip.Addr
}

func ResolveIp46(ctx context.Context, dialer netproxy.Dialer, dns netip.AddrPort, host string, tcp bool, race bool) (ipv46 *Ip46, err error) {
	var wg sync.WaitGroup
	wg.Add(2)
	var err4, err6 error
	var addrs4, addrs6 []netip.Addr
	ctx4, cancel4 := context.WithCancel(ctx)
	ctx6, cancel6 := context.WithCancel(ctx)
	go func() {
		defer func() {
			wg.Done()
			cancel4()
			if race {
				cancel6()
			}
		}()
		var e error
		addrs4, e = ResolveNetip(ctx4, dialer, dns, host, dnsmessage.TypeA, tcp)
		if err != nil && !errors.Is(e, context.Canceled) {
			err4 = e
			return
		}
		if len(addrs4) == 0 {
			addrs4 = []netip.Addr{{}}
		}
	}()
	go func() {
		defer func() {
			wg.Done()
			cancel6()
			if race {
				cancel4()
			}
		}()
		var e error
		addrs6, e = ResolveNetip(ctx6, dialer, dns, host, dnsmessage.TypeAAAA, tcp)
		if err != nil && !errors.Is(e, context.Canceled) {
			err6 = e
			return
		}
		if len(addrs6) == 0 {
			addrs6 = []netip.Addr{{}}
		}
	}()
	wg.Wait()
	if err4 != nil || err6 != nil {
		if err4 != nil && err6 != nil {
			return nil, fmt.Errorf("%w: %v", err4, err6)
		}
		if err4 != nil {
			return nil, err4
		} else {
			return nil, err6
		}
	}
	return &Ip46{
		Ip4: addrs4[0],
		Ip6: addrs6[0],
	}, nil
}
