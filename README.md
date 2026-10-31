# FixLayout Go

**English** | [ภาษาไทย](#ภาษาไทย)

FixLayout fixes text typed with the wrong keyboard layout on Linux **X11/Xorg**. For example, it can convert Thai ↔ US English characters based on the physical keyboard keys. The project is written in Go and uses `libxkbcommon`, `xclip`, and `xdotool`.

## Features

- **Whole selection / line:** Convert selected text, or text from the start of the current line to the cursor if nothing is selected. Automatic layout choice uses available Hunspell dictionaries or a fallback Thai plausibility check.
- **One character at a time (`--char`):** Convert the character immediately to the left of the cursor (Thai ↔ US English), then **move the cursor one position left**, including when the character is unchanged. Repeat the shortcut to work backwards.
- Explicit conversion using `--force us:th` or `--force th:us`.
- Detect configured layouts with `setxkbmap`, or specify `--layouts us,th,de`.
- Restore the previous clipboard after pasting (with a delay to allow target applications to read it).
- Use `Ctrl+C` / `Ctrl+V` by default rather than Insert-based shortcuts, which can affect insert/overwrite mode in some applications.

## Requirements

Linux X11/Xorg (not native Wayland); Go 1.18+ and GCC for building with cgo; `libxkbcommon0`, `xkb-data`, `xclip`, and `xdotool` at runtime. `setxkbmap` is used for layout discovery (fallback: `us,th`). Recommended optional tools: `libnotify-bin`, `hunspell-en-us`, and `hunspell-th`.

## Build and install (Ubuntu / Linux Mint)

```bash
sudo apt-get update
sudo apt-get install -y golang-go build-essential libxkbcommon0 xkb-data \
  xclip xdotool libnotify-bin hunspell-en-us
./build.sh
./install.sh --no-deps
```

`build.sh` produces `bin/fixlayout-go`. `install.sh` builds and installs it as `~/.local/bin/fixlayout` and configures GNOME/Cinnamon shortcuts using `gsettings`. By default, **Ctrl+Alt+K** fixes a selection/line and **Ctrl+Backspace** fixes one character. This overrides the usual delete-previous-word action. The installer backs up a pre-existing executable as `~/.local/bin/fixlayout.before-go.bak` if that backup does not yet exist.

For a fresh machine, `./install.sh` installs the dependencies too. Customize shortcuts:

```bash
KEY='<Primary><Alt>j' CHAR_KEY='<Primary>BackSpace' ./install.sh --no-deps
# Disable the character shortcut:
CHAR_KEY=none ./install.sh --no-deps
```

## CLI examples

```bash
./bin/fixlayout-go --char --text 'l'    # ส
./bin/fixlayout-go --char --text 'ส'    # l
./bin/fixlayout-go --force us:th --text hello
./bin/fixlayout-go --list
./bin/fixlayout-go --char              # edit in the focused app
./bin/fixlayout-go --char --clipboard-keys=insert # legacy copy/paste shortcuts
```

## Tests

```bash
go test ./...
```

## Limitations

- Requires X11/Xorg. Native Wayland is not supported.
- Character mode is not supported in terminals; terminal copy/paste shortcuts differ.
- Focused applications must support standard text selection, copy, and paste. Browser text fields are the primary use case; behavior may differ in LibreOffice, Gedit, or Xed.
- `--clipboard-keys=insert` restores `Ctrl+Insert` / `Shift+Insert` if necessary, but can affect cursor mode in some applications.
- The program does **not** explicitly switch between insert and overwrite mode; it just avoids sending Insert in the default mode.
- Cursor movement is one application-defined position, which may treat Thai combining marks as a group.
- Clipboard changes from other processes may race with the program; Go binaries built on other machines may have different glibc compatibility.

---

# ภาษาไทย

FixLayout Go เป็นโปรแกรมแก้ข้อความที่พิมพ์ผิดภาษาแป้นพิมพ์บน Linux **X11/Xorg** เช่น พิมพ์โดยลืมเปลี่ยนภาษาไทย ↔ อังกฤษ โปรแกรมแปลงอักษรตามตำแหน่งปุ่มจริง ใช้ภาษา Go ร่วมกับ `libxkbcommon`, `xclip` และ `xdotool`

## ความสามารถ

- **แก้ข้อความทั้งชุด:** แปลงข้อความที่เลือกไว้ หรือข้อความตั้งแต่ต้นบรรทัดจนถึงเคอร์เซอร์เมื่อไม่ได้เลือกข้อความ โปรแกรมเลือกทิศทางจาก dictionary Hunspell (หรือใช้ตัวประเมินภาษาไทยเมื่อไม่มี dictionary)
- **แก้ทีละตัว (`--char`):** แปลงตัวอักษรทางซ้ายของเคอร์เซอร์จากไทย ↔ อังกฤษ แล้ว **เลื่อนเคอร์เซอร์ถอยซ้าย 1 ตำแหน่ง** แม้ตัวนั้นจะไม่เปลี่ยน กดปุ่มลัดซ้ำเพื่อแก้ย้อนกลับได้
- ใช้ `--force us:th` หรือ `--force th:us` เพื่อบังคับทิศทาง
- อ่าน layout จาก `setxkbmap` หรือกำหนดเองด้วย `--layouts us,th,de`
- คืน Clipboard เดิมหลังวางข้อความ โดยมีเวลารอให้แอปปลายทางรับข้อความ
- ใช้ `Ctrl+C` / `Ctrl+V` เป็นค่าเริ่มต้น ไม่ส่งปุ่ม Insert เพื่อลดปัญหาโหมด Insert/Overwrite

## สิ่งที่ต้องมี

Linux แบบ X11/Xorg (ยังไม่รองรับ Wayland โดยตรง), Go 1.18+ และ GCC สำหรับคอมไพล์, `libxkbcommon0`, `xkb-data`, `xclip`, `xdotool` ส่วน `setxkbmap` ใช้อ่าน layout (ถ้าไม่มีจะใช้ `us,th`) แนะนำเพิ่มเติม: `libnotify-bin`, `hunspell-en-us`, `hunspell-th`

## คอมไพล์และติดตั้ง (Ubuntu / Linux Mint)

```bash
sudo apt-get update
sudo apt-get install -y golang-go build-essential libxkbcommon0 xkb-data \
  xclip xdotool libnotify-bin hunspell-en-us
./build.sh
./install.sh --no-deps
```

คอมไพล์แล้วได้ไฟล์ `bin/fixlayout-go` ส่วน `install.sh` จะติดตั้งไปที่ `~/.local/bin/fixlayout` และตั้งปุ่มลัดอัตโนมัติบน GNOME/Cinnamon ผ่าน `gsettings` โดยค่าเริ่มต้น **Ctrl+Alt+K** แก้ทั้งข้อความ และ **Ctrl+Backspace** แก้ทีละตัว (จะแทนที่คำสั่งลบทีละคำของแอป) หากมีโปรแกรมเดิมอยู่ จะสำรองเป็น `~/.local/bin/fixlayout.before-go.bak` ถ้ายังไม่มีไฟล์สำรอง

หากต้องการติดตั้ง dependencies ด้วย ให้ใช้ `./install.sh` แทน สามารถเปลี่ยนปุ่มลัดได้ เช่น

```bash
KEY='<Primary><Alt>j' CHAR_KEY='<Primary>BackSpace' ./install.sh --no-deps
# ไม่ตั้งปุ่มลัดแก้ทีละตัว:
CHAR_KEY=none ./install.sh --no-deps
```

## ตัวอย่างคำสั่ง

```bash
./bin/fixlayout-go --char --text 'l'    # ส
./bin/fixlayout-go --char --text 'ส'    # l
./bin/fixlayout-go --force us:th --text hello
./bin/fixlayout-go --list
./bin/fixlayout-go --char
./bin/fixlayout-go --char --clipboard-keys=insert
```

## ทดสอบ

```bash
go test ./...
```

## ข้อจำกัด

- รองรับ X11/Xorg ไม่รองรับ native Wayland
- โหมด `--char` ไม่รองรับ Terminal เพราะการคัดลอก/วางใช้ปุ่มต่างจากแอปทั่วไป
- แอปที่โฟกัสต้องรองรับการเลือกข้อความและ Copy/Paste โดยใช้ Browser เป็นหลัก ส่วน LibreOffice, Gedit และ Xed อาจมีพฤติกรรมต่างกัน
- หาก `Ctrl+C` / `Ctrl+V` ใช้งานไม่ได้ สามารถตั้ง `--clipboard-keys=insert` เพื่อใช้ `Ctrl+Insert` / `Shift+Insert` แบบเดิม แต่บางแอปอาจเปลี่ยนสถานะเคอร์เซอร์
- โปรแกรมไม่ได้สลับโหมด Insert/Overwrite เอง เพียงหลีกเลี่ยงการส่งปุ่ม Insert ในค่าเริ่มต้น
- การเลื่อนซ้ายหนึ่งตำแหน่งขึ้นกับแอป ตัวอักษรผสมภาษาไทยอาจถูกมองเป็นกลุ่ม
- Clipboard เป็นทรัพยากรร่วม อาจชนกับโปรแกรมอื่น และ Binary ที่คอมไพล์ต่างเครื่องอาจมีปัญหาความเข้ากันได้ของ glibc