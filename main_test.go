package main
import "testing"
func TestLayoutsConvertRoundTrip(t *testing.T){
x,err:=newXKB();if err!=nil{t.Fatal(err)};defer x.Close()
original:="hello World"
converted,err:=x.Convert(original,"us","th");if err!=nil{t.Fatal(err)}
if converted==original{t.Fatalf("conversion unchanged")}
restored,err:=x.Convert(converted,"th","us");if err!=nil{t.Fatal(err)}
if restored!=original{t.Fatalf("roundtrip %q != %q",restored,original)}
}
func TestCharFix(t *testing.T){
x,err:=newXKB();if err!=nil{t.Fatal(err)};defer x.Close()
got,err:=x.CharFix("l");if err!=nil{t.Fatal(err)};if got!="ส"{t.Fatalf("expected ส, got %q",got)}
got,err=x.CharFix("ส");if err!=nil{t.Fatal(err)};if got!="l"{t.Fatalf("expected l, got %q",got)}
}
func TestThaiPlausible(t *testing.T){if thaiPlausible("ภาษาไทย")<=thaiPlausible("้่ิ"){t.Fatal("unexpected Thai plausibility")}}
func TestChooseWithDictionary(t *testing.T){
x,err:=newXKB();if err!=nil{t.Fatal(err)};defer x.Close()
d:=newDictionary();d.cache["en"]=map[string]bool{"hello":true};d.cache["th"]=map[string]bool{}
thai,err:=x.Convert("hello","us","th");if err!=nil{t.Fatal(err)}
result,err:=choose(x,d,thai,[]string{"us","th"});if err!=nil{t.Fatal(err)}
if result==nil||result.text!="hello"{t.Fatalf("expected hello, got %#v",result)}
}
func TestClipboardShortcutsAvoidInsertByDefault(t *testing.T){
copy,paste,err:=clipboardShortcuts(false,"ctrl");if err!=nil||copy!="ctrl+c"||paste!="ctrl+v"{t.Fatalf("invalid GUI shortcuts")}
copy,paste,err=clipboardShortcuts(true,"ctrl");if err!=nil||copy!="ctrl+shift+c"||paste!="ctrl+shift+v"{t.Fatalf("invalid terminal shortcuts")}
}
func TestLegacyInsertClipboardShortcutsRemainAvailable(t *testing.T){
copy,paste,err:=clipboardShortcuts(false,"insert");if err!=nil||copy!="ctrl+Insert"||paste!="shift+Insert"{t.Fatalf("invalid legacy shortcuts")}
if _,_,err:=clipboardShortcuts(false,"invalid");err==nil{t.Fatal("expected validation error")}
}
