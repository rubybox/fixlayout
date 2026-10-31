//go:build linux && cgo

package main

/*
#cgo LDFLAGS: -l:libxkbcommon.so.0
#include <stdint.h>
#include <stdlib.h>
#include <stddef.h>
struct xkb_context;
struct xkb_keymap;
struct xkb_state;
struct xkb_rule_names {
    const char *rules;
    const char *model;
    const char *layout;
    const char *variant;
    const char *options;
};
struct xkb_context *xkb_context_new(int flags);
void xkb_context_unref(struct xkb_context *context);
struct xkb_keymap *xkb_keymap_new_from_names(struct xkb_context *context,
    const struct xkb_rule_names *names, int flags);
void xkb_keymap_unref(struct xkb_keymap *keymap);
struct xkb_state *xkb_state_new(struct xkb_keymap *keymap);
void xkb_state_unref(struct xkb_state *state);
uint32_t xkb_keymap_mod_get_index(struct xkb_keymap *keymap, const char *name);
uint32_t xkb_state_update_mask(struct xkb_state *state,
    uint32_t depressed, uint32_t latched, uint32_t locked,
    uint32_t depressed_layout, uint32_t latched_layout, uint32_t locked_layout);
int xkb_state_key_get_utf8(struct xkb_state *state, uint32_t key,
    char *buffer, size_t size);
*/
import "C"

import (
    "fmt"
    "strings"
    "unicode"
    "unicode/utf8"
    "unsafe"
)
type position struct { keycode uint32; shift bool }
type keyTable struct { reverse map[rune]position; forward map[position]rune }
type XKB struct { ctx *C.struct_xkb_context; tables map[string]*keyTable }
func newXKB() (*XKB,error) {
    ctx:=C.xkb_context_new(0)
    if ctx==nil {return nil,fmt.Errorf("cannot open libxkbcommon")}
    return &XKB{ctx:ctx,tables:make(map[string]*keyTable)},nil
}
func (x *XKB) Close() {if x.ctx!=nil {C.xkb_context_unref(x.ctx);x.ctx=nil}}
func parseSpec(spec string)(string,string) {
    spec=strings.TrimSpace(spec)
    left:=strings.IndexByte(spec,'(')
    if left>=0 && strings.HasSuffix(spec,")") {return strings.TrimSpace(spec[:left]),spec[left+1:len(spec)-1]}
    return spec,""
}
func (x *XKB) table(spec string)(*keyTable,error) {
    if t,ok:=x.tables[spec];ok{return t,nil}
    name,variant:=parseSpec(spec)
    if name=="" {return nil,fmt.Errorf("empty layout")}
    cname:=C.CString(name);defer C.free(unsafe.Pointer(cname))
    names:=C.struct_xkb_rule_names{layout:cname}
    if variant!="" {v:=C.CString(variant);defer C.free(unsafe.Pointer(v));names.variant=v}
    km:=C.xkb_keymap_new_from_names(x.ctx,&names,0)
    if km==nil {return nil,fmt.Errorf("could not load layout %q",spec)}
    defer C.xkb_keymap_unref(km)
    state:=C.xkb_state_new(km)
    if state==nil{return nil,fmt.Errorf("could not create state for %q",spec)}
    defer C.xkb_state_unref(state)
    sh:=C.CString("Shift");defer C.free(unsafe.Pointer(sh))
    shiftIndex:=uint32(C.xkb_keymap_mod_get_index(km,sh))
    if shiftIndex>=32{return nil,fmt.Errorf("no Shift modifier for %q",spec)}
    t:=&keyTable{reverse:map[rune]position{},forward:map[position]rune{}}
    for kc:=uint32(9);kc<100;kc++ {
        for _,shift:=range []bool{false,true} {
            var mask uint32
            if shift {mask=1<<shiftIndex}
            C.xkb_state_update_mask(state,C.uint32_t(mask),0,0,0,0,0)
            var buf [32]C.char
            n:=C.xkb_state_key_get_utf8(state,C.uint32_t(kc),&buf[0],C.size_t(len(buf)))
            if n<=0||int(n)>=len(buf){continue}
            value:=C.GoString(&buf[0])
            if utf8.RuneCountInString(value)!=1{continue}
            r,_:=utf8.DecodeRuneInString(value)
            if !unicode.IsPrint(r){continue}
            p:=position{kc,shift}
            t.forward[p]=r
            if _,exists:=t.reverse[r];!exists {t.reverse[r]=p}
        }
    }
    x.tables[spec]=t
    return t,nil
}
func (x *XKB) Convert(input,src,dst string)(string,error) {
    from,err:=x.table(src);if err!=nil{return "",err}
    to,err:=x.table(dst);if err!=nil{return "",err}
    var out strings.Builder
    out.Grow(len(input))
    for _,r:=range input {
        if p,found:=from.reverse[r];found {
            if mapped,ok:=to.forward[p];ok{out.WriteRune(mapped);continue}
        }
        out.WriteRune(r)
    }
    return out.String(),nil
}
func (x *XKB) Coverage(input,src string)(float64,error){
    t,err:=x.table(src);if err!=nil{return 0,err}
    total,covered:=0,0
    for _,r:=range input {
        if unicode.IsSpace(r){continue}
        total++
        if _,ok:=t.reverse[r];ok{covered++}
    }
    if total==0{return 0,nil}
    return float64(covered)/float64(total),nil
}
func (x *XKB) CharFix(input string)(string,error){
    var out strings.Builder
    for _,r:=range input {
        var converted string
        var err error
        if r>='\u0e00' && r<='\u0e7f'{converted,err=x.Convert(string(r),"th","us")}else{converted,err=x.Convert(string(r),"us","th")}
        if err!=nil{return "",err}
        out.WriteString(converted)
    }
    return out.String(),nil
}
