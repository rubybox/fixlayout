package main
import (
"bufio"
"errors"
"flag"
"fmt"
"os"
"os/exec"
"path/filepath"
"strings"
"unicode/utf8"
)
var langForLayout=map[string]string{"us":"en","gb":"en","au":"en","ca":"en","in":"en","ie":"en","nz":"en","za":"en","jp":"ja","kr":"ko","cn":"zh","br":"pt","latam":"es","ua":"uk","gr":"el","il":"he","cz":"cs","dk":"da","se":"sv","no":"nb","ara":"ar","ir":"fa","la":"lo","kh":"km"}
var unspacedLang=map[string]bool{"th":true,"lo":true,"km":true,"my":true,"ja":true,"zh":true}
var suffixes=[]string{"s","es","ed","ing","ly","'s","d","er"}
const punctuation = `!"#$%&'()*+,-./:;<=>?@[\\]^_` + "`" + `{|}~“”‘’…`
type Dictionary struct{cache map[string]map[string]bool}
func newDictionary()*Dictionary{return &Dictionary{cache:map[string]map[string]bool{}}}
func (d *Dictionary) words(spec string)map[string]bool{
name,_:=parseSpec(spec);lang:=langForLayout[name];if lang==""{lang=name}
if words,ok:=d.cache[lang];ok{return words}
words:=map[string]bool{}
locations:=[]string{"/usr/share/hunspell","/usr/share/myspell","/usr/share/myspell/dicts","/usr/share/dict","/usr/share/libreoffice/share/extensions/*","/usr/lib/libreoffice/share/extensions/*"}
if home,err:=os.UserHomeDir();err==nil{locations=append(locations,filepath.Join(home,".config/fixlayout"))}
seen:=map[string]bool{}
for _,pattern:=range locations{
roots,_:=filepath.Glob(pattern)
for _,root:=range roots {
for _,part:=range []string{lang+"_*.dic",lang+".dic","words-"+lang+".txt"}{
paths,_:=filepath.Glob(filepath.Join(root,part))
for _,file:=range paths{if !seen[file]{seen[file]=true;readDictionary(file,words)}}
}}}
if lang=="en"{for _,path:=range []string{"/usr/share/dict/american-english","/usr/share/dict/words"}{if !seen[path]{readDictionary(path,words)}}}
d.cache[lang]=words;return words
}
func readDictionary(path string,words map[string]bool){
file,err:=os.Open(path);if err!=nil{return};defer file.Close()
scan:=bufio.NewScanner(file);scan.Buffer(make([]byte,4096),2*1024*1024)
first:=true
for scan.Scan(){
line:=scan.Text()
if first{first=false;var count int;if _,err:=fmt.Sscanf(strings.TrimSpace(line),"%d",&count);err==nil&&strings.TrimSpace(line)==fmt.Sprint(count){continue}}
word:=strings.ToLower(strings.TrimSpace(strings.SplitN(line,"/",2)[0]));if word!=""{words[word]=true}
}}
func known(word string,words map[string]bool)bool{
if words[word]{return true}
for _,suffix:=range suffixes{if strings.HasSuffix(word,suffix)&&utf8.RuneCountInString(word)>len(suffix)+2&&words[strings.TrimSuffix(word,suffix)]{return true}}
return false
}
func segmentCover(token string,words map[string]bool)int{
rs:=[]rune(token);dp:=make([]int,len(rs)+1)
for i:=1;i<=len(rs);i++{dp[i]=dp[i-1];from:=i-15;if from<0{from=0};for j:=from;j<i-1;j++{if words[string(rs[j:i])]&&dp[j]+i-j>dp[i]{dp[i]=dp[j]+i-j}}}
return dp[len(rs)]
}
const thMarks="ัิีึืุู็่้๊๋์"
const thTones="่้๊๋์็"
const thNoLead="ะัาำิีึืฺุู็่้๊๋์ๆฯ"
const thNoEnd="เแโใไ"
func thConsonant(r rune)bool{return r>=0x0e01&&r<=0x0e2e}
func thaiPlausible(tok string)float64{
rs:=[]rune(tok);if len(rs)==0{return 0}
hasConsonant:=false;for _,r:=range rs{hasConsonant=hasConsonant||thConsonant(r)}
if !hasConsonant{return 0}
bad:=0;if strings.ContainsRune(thNoLead,rs[0]){bad+=2}
if strings.ContainsRune(thNoEnd,rs[len(rs)-1]){bad++}
for i:=1;i<len(rs);i++{c,prev:=rs[i],rs[i-1];if strings.ContainsRune(thMarks,c)&&!thConsonant(prev)&&!(strings.ContainsRune(thTones,c)&&strings.ContainsRune(thMarks,prev)){bad++}}
denom:=len(rs);if denom<2{denom=2}
score:=1-float64(bad)/float64(denom);if score<0{score=0};return score*0.7
}
func (d *Dictionary) score(text,spec string)float64{
name,_:=parseSpec(spec);lang:=langForLayout[name];if lang==""{lang=name}
tokens:=make([]string,0);total:=0
for _,field:=range strings.Fields(strings.ToLower(text)){tok:=strings.Trim(field,punctuation);if tok!=""{tokens=append(tokens,tok);total+=utf8.RuneCountInString(tok)}}
if total==0{return 0}
words:=d.words(spec)
if len(words)>0{
hit:=0
for _,tok:=range tokens{if unspacedLang[lang]{hit+=segmentCover(tok,words)}else if known(tok,words){hit+=utf8.RuneCountInString(tok)}}
return float64(hit)/float64(total)
}
if name=="th"{hit:=0.0;for _,tok:=range tokens{hit+=thaiPlausible(tok)*float64(utf8.RuneCountInString(tok))};return hit/float64(total)}
return 0
}
type decision struct{text,src,dst string;confidence float64}
func choose(x *XKB,dict *Dictionary,input string,layouts []string)(*decision,error){
sources:=make([]string,0,len(layouts))
for _,layout:=range layouts{coverage,err:=x.Coverage(input,layout);if err!=nil{return nil,err};if coverage>=0.9{sources=append(sources,layout)}}
if len(sources)==0{sources=layouts}
var best *decision;asIs:=0.0
for _,src:=range sources{
if value:=dict.score(input,src);value>asIs{asIs=value}
for _,dst:=range layouts{
if dst==src{continue}
converted,err:=x.Convert(input,src,dst);if err!=nil{return nil,err};if converted==input{continue}
confidence:=dict.score(converted,dst)
if best==nil||confidence>best.confidence{best=&decision{converted,src,dst,confidence}}
}}
if best!=nil&&asIs>=0.6&&asIs>=best.confidence{return nil,nil}
return best,nil
}
func detectLayouts()[]string{
output,err:=exec.Command("setxkbmap","-query").Output();if err!=nil{return []string{"us","th"}}
layout,variant:="",""
for _,line:=range strings.Split(string(output),"\n"){
field:=strings.SplitN(line,":",2);if len(field)!=2{continue}
switch strings.TrimSpace(field[0]){case "layout":layout=strings.TrimSpace(field[1]);case "variant":variant=strings.TrimSpace(field[1])}
}
names:=strings.Split(layout,",");variants:=strings.Split(variant,",");layouts:=make([]string,0,len(names))
for i,name:=range names{name=strings.TrimSpace(name);if name==""{continue};if i<len(variants)&&strings.TrimSpace(variants[i])!=""{name+="("+strings.TrimSpace(variants[i])+")"};layouts=append(layouts,name)}
if len(layouts)>=2{return layouts};return []string{"us","th"}
}
func run()error{
var layoutsOpt,forceOpt,textOpt,switchOpt,clipboardOpt string
var charMode,list bool
flag.StringVar(&layoutsOpt,"layouts","","keyboard layouts e.g. us,th,ru")
flag.StringVar(&forceOpt,"force","","direction e.g. us:th")
flag.StringVar(&textOpt,"text","","convert CLI text without clipboard")
flag.StringVar(&switchOpt,"switch-key","","keyboard language switch key e.g. super+space")
flag.StringVar(&clipboardOpt,"clipboard-keys","ctrl","ctrl (default) or insert")
flag.BoolVar(&charMode,"char",false,"fix one character then move cursor left")
flag.BoolVar(&list,"list",false,"list layouts and dictionary sizes")
flag.Parse()
hasText:=false;flag.Visit(func(f *flag.Flag){if f.Name=="text"{hasText=true}})
layouts:=detectLayouts()
if layoutsOpt!=""{layouts=nil;for _,l:=range strings.Split(layoutsOpt,","){if s:=strings.TrimSpace(l);s!=""{layouts=append(layouts,s)}}}
if charMode{layouts=[]string{"us","th"}}
if len(layouts)==0{return errors.New("no keyboard layout provided")}
src,dst:="",""
if forceOpt!=""{pair:=strings.Split(forceOpt,":");if len(pair)!=2||pair[0]==""||pair[1]==""{return errors.New("--force must be src:dst e.g. us:th")};src,dst=pair[0],pair[1]}
x,err:=newXKB();if err!=nil{return err};defer x.Close()
dict:=newDictionary()
if list{for _,l:=range layouts{if _,err:=x.table(l);err!=nil{return err};fmt.Printf("%s: dictionary %d words\n",l,len(dict.words(l)))};return nil}
if hasText{
if charMode{result,err:=x.CharFix(textOpt);if err!=nil{return err};fmt.Println(result);return nil}
if src!=""{result,err:=x.Convert(textOpt,src,dst);if err!=nil{return err};fmt.Println(result);return nil}
result,err:=choose(x,dict,textOpt,layouts);if err!=nil{return err}
if result==nil{fmt.Println(textOpt)}else{fmt.Println(result.text);fmt.Fprintf(os.Stderr,"# %s -> %s score %.2f\n",result.src,result.dst,result.confidence)}
return nil
}
return runHotkey(x,dict,layouts,src,dst,switchOpt,clipboardOpt,charMode)
}
func main(){if err:=run();err!=nil{fmt.Fprintln(os.Stderr,"fixlayout:",err);os.Exit(1)}}
