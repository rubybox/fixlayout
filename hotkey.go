package main
import (
"context"
"fmt"
"os"
"os/exec"
"path/filepath"
"strings"
"syscall"
"time"
)
const (
releaseDelay=110*time.Millisecond
pasteDelay=100*time.Millisecond
clipboardPoll=15*time.Millisecond
)
func checkCommand(name string) error { if _,err:=exec.LookPath(name);err!=nil{return fmt.Errorf("missing %s: sudo apt install %s",name,name)};return nil }
func getClip()(string,error){ctx,cancel:=context.WithTimeout(context.Background(),2*time.Second);defer cancel();v,e:=exec.CommandContext(ctx,"xclip","-selection","clipboard","-o").Output();return string(v),e}
func setClip(value string)error{ctx,cancel:=context.WithTimeout(context.Background(),2*time.Second);defer cancel();cmd:=exec.CommandContext(ctx,"xclip","-selection","clipboard","-i");cmd.Stdin=strings.NewReader(value);return cmd.Run()}
func key(combo string)error{ctx,cancel:=context.WithTimeout(context.Background(),2*time.Second);defer cancel();return exec.CommandContext(ctx,"xdotool","key","--clearmodifiers","--delay","0",combo).Run()}
func notify(message string){if _,err:=exec.LookPath("notify-send");err==nil{_ = exec.Command("notify-send","-t","1600","fixlayout",message).Run()}}
func isTerminal()bool{
output,err:=exec.Command("xdotool","getwindowfocus","getwindowclassname").Output();if err!=nil{return false}
lower:=strings.ToLower(strings.TrimSpace(string(output)))
for _,name:=range []string{"terminal","konsole","xterm","tilix","kitty","alacritty","terminator","urxvt","guake","tmux","wezterm","foot"}{if strings.Contains(lower,name){return true}}
return false
}
func waitForClipboard(match func(string)bool,timeout time.Duration)(string,bool){
deadline:=time.Now().Add(timeout)
for {value,err:=getClip();if err==nil&&match(value){return value,true};if time.Now().After(deadline){return "",false};time.Sleep(clipboardPoll)}
}
func lockHotkey()(func(),error){
dir,err:=os.UserCacheDir();if err!=nil{return nil,err};dir=filepath.Join(dir,"fixlayout-go")
if err:=os.MkdirAll(dir,0700);err!=nil{return nil,err}
file,err:=os.OpenFile(filepath.Join(dir,"hotkey.lock"),os.O_CREATE|os.O_RDWR,0600);if err!=nil{return nil,err}
if err:=syscall.Flock(int(file.Fd()),syscall.LOCK_EX);err!=nil{_ = file.Close();return nil,err}
return func(){_ = syscall.Flock(int(file.Fd()),syscall.LOCK_UN);_ = file.Close()},nil
}
func clipboardShortcuts(terminal bool,mode string)(copyCombo,pasteCombo string,err error){
switch mode {
case "ctrl":if terminal{return "ctrl+shift+c","ctrl+shift+v",nil};return "ctrl+c","ctrl+v",nil
case "insert":if terminal{return "ctrl+shift+c","ctrl+shift+v",nil};return "ctrl+Insert","shift+Insert",nil
default:return "","",fmt.Errorf("--clipboard-keys must be ctrl or insert")
}}
func runHotkey(x *XKB,dict *Dictionary,layouts []string,src,dst,switchKey,clipboardMode string,charMode bool)error{
for _,cmd:=range []string{"xclip","xdotool"}{if err:=checkCommand(cmd);err!=nil{return err}}
if os.Getenv("XDG_SESSION_TYPE")=="wayland"{notify("X11/Xorg required");return fmt.Errorf("Wayland is unsupported")}
if os.Getenv("DISPLAY")==""{return fmt.Errorf("DISPLAY not set; X11/Xorg required")}
unlock,err:=lockHotkey();if err!=nil{return err};defer unlock()
terminal:=isTerminal();if terminal&&charMode{notify("Character mode not supported in terminals");return nil}
copyCombo,pasteCombo,err:=clipboardShortcuts(terminal,clipboardMode);if err!=nil{return err}
time.Sleep(releaseDelay)
old,oldErr:=getClip();if oldErr!=nil{old=""}
defer func(){if err:=setClip(old);err!=nil{fmt.Fprintln(os.Stderr,"fixlayout: clipboard restore failed:",err)}}()
sentinel:=fmt.Sprintf("__fixlayout_go_%d_%d__",os.Getpid(),time.Now().UnixNano())
if err:=setClip(sentinel);err!=nil{return err}
if _,ok:=waitForClipboard(func(s string)bool{return s==sentinel},550*time.Millisecond);!ok{return fmt.Errorf("clipboard not ready")}
if err:=key(copyCombo);err!=nil{return err}
text,found:=waitForClipboard(func(s string)bool{return s!=sentinel},180*time.Millisecond)
auto:=!found
if auto{
sel:="shift+Home";if charMode{sel="shift+Left"}
if err:=key(sel);err!=nil{return err}
if err:=key(copyCombo);err!=nil{return err}
text,found=waitForClipboard(func(s string)bool{return s!=sentinel},350*time.Millisecond)
}
if !found||text==""{if charMode&&auto{_ = key("Left")};notify("No text selected");return nil}
var replacement string;shouldReplace:=true
if charMode{replacement,err=x.CharFix(text);shouldReplace=replacement!=text
}else if src!=""{replacement,err=x.Convert(text,src,dst)
}else{result,chooseErr:=choose(x,dict,text,layouts);err=chooseErr;shouldReplace=result!=nil;if result!=nil{replacement=result.text}}
if err!=nil{return err}
if !shouldReplace{if charMode{return key("Left")};notify("Text appears correct");return nil}
if err:=setClip(replacement);err!=nil{return err}
if _,ok:=waitForClipboard(func(s string)bool{return s==replacement},300*time.Millisecond);!ok{return fmt.Errorf("replacement clipboard not ready")}
if err:=key(pasteCombo);err!=nil{return err}
time.Sleep(pasteDelay)
if charMode{if err:=key("Left");err!=nil{return err}}
if switchKey!=""{return key(switchKey)}
return nil
}
