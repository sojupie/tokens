package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/a-h/templ"
	"github.com/sojupie/tokens/forge/components/ui"
)

// This page uses only the public Forge API, without Catalog adapters or scripts.
func apiContracts(w http.ResponseWriter, r *http.Request) {
	assets := ui.Assets{}
	ctx := ui.WithAssets(r.Context(), assets)
	var content bytes.Buffer
	results := templ.ComponentFunc(func(ctx context.Context, out io.Writer) error {
		for _, label := range []string{"Alpha", "Beta"} {
			if err := ui.SearchResult(label).Render(ctx, out); err != nil {
				return err
			}
		}
		return nil
	})
	if err := ui.Search(ui.SearchProps{ID: "standalone-search", Label: "Find records", Leading: "icon", Variant: "docked", Actions: "avatar-clear"}).Render(templ.WithChildren(ctx, results), &content); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	destinations := templ.ComponentFunc(func(ctx context.Context, out io.Writer) error {
		for i, label := range []string{"Inbox", "Saved", "Settings"} {
			if err := ui.NavigationItem(ui.NavigationItemProps{Label: label, Icon: []string{"mail", "star", "settings"}[i], Current: i == 0, Badge: map[bool]string{true: "small"}[i == 1], BadgeLabel: "Unread"}).Render(ctx, out); err != nil {
				return err
			}
		}
		return nil
	})
	if err := ui.NavigationRail(ui.RailProps{NavigationProps: ui.NavigationProps{ID: "standalone-rail", Label: "Mail", Variant: "collapsed", Width: "default"}, FAB: &ui.FABProps{Label: "Compose", Icon: "edit"}}).Render(templ.WithChildren(ctx, destinations), &content); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!doctype html><meta charset="utf-8"><title>Standalone component API checks</title><link rel="stylesheet" href="/themes/shared.css"><link rel="stylesheet" href="/themes/common.css"><link rel="stylesheet" href="/static/buttons.css"><link rel="stylesheet" href="/static/bundles/common.css"><style>body{margin:24px;font:16px system-ui}.md-search-demo{height:420px}#standalone-rail{height:440px}pre{white-space:pre-wrap}</style>`)
	names := make([]string, 0, len(assets))
	for name := range assets {
		if name != "common" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(w, `<link rel="stylesheet" href="/themes/%s.css"><link rel="stylesheet" href="/static/bundles/%s.css">`, name, name)
	}
	fmt.Fprint(w, `<h1>Standalone component API checks</h1><button id="run">Run API checks</button><pre id="result">Ready</pre>`)
	w.Write(content.Bytes())
	fmt.Fprint(w, `<script src="/static/motion.js?v=api-boundary-20261102"></script><script src="/static/buttons.js?v=api-boundary-20261102"></script><script src="/static/runtime/utils.js?v=api-boundary-20261102"></script><script src="/static/runtime/navigation.js?v=api-boundary-20261102"></script><script src="/static/search.js?v=api-boundary-20261102"></script><script>
 const errors=[];window.addEventListener('error',e=>errors.push(e.message));
 document.querySelector('#run').onclick=async()=>{
  const checks=[],assert=(yes,message)=>{if(!yes)throw Error(message);checks.push(message);};
  try{
   const root=document.querySelector('.md-search-demo'),input=root.querySelector('input'),profile=root.querySelector('.md-search-avatar');let selected,profiles=0;
   root.addEventListener('md-search-select',e=>selected=e.detail.label,{once:true});root.addEventListener('md-search-profile',()=>profiles++);
   profile.click();assert(profiles===1,'Profile action emits an application event without feedback markup');
   input.focus();assert(root.querySelector('.md-search-view').dataset.open==='true','Search opens without a Catalog page');
   root.querySelector('[data-search-result] button').click();
   assert(selected==='Alpha','Selection emits the chosen result without sample headings');
   assert(root.querySelector('.md-search-view').dataset.open==='false','Selection closes search');
   assert(document.activeElement===input,'Search restores focus to its own input');
   const rail=document.querySelector('#standalone-rail'),saved=rail.querySelectorAll('[data-navigation-item]')[1],badge=saved.querySelector('.md-badge');
   saved.click();assert(saved.getAttribute('aria-current')==='page','Typed navigation item changes selection');
   assert(saved.querySelector('.md-badge')===badge&&badge,'Navigation leaves unread state to the application');
   assert(getComputedStyle(rail.querySelector('.md-nav-fab')).getPropertyValue('--_elevation-level').trim()==='0','Typed nested FAB uses rail elevation');
   assert(errors.length===0,'No application markup dependencies throw at runtime');
   assert(!document.querySelector('[data-component-case],[data-component-state],#component-preview'),'No Catalog adapter or playground state is present');
   document.querySelector('#result').textContent=JSON.stringify({passed:checks.length,checks},null,2);
  }catch(error){document.querySelector('#result').textContent=JSON.stringify({passed:checks.length,checks,error:error.stack,errors},null,2);}
 };</script>`)
}
