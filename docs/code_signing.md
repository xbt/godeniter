# Windows 数字签名与代码证书实战手册 (Authenticode Code Signing)

本文档介绍如何在 **Godeniter** 框架及脚手架开发完成后，使用 **纯 Go 标准库 0 依赖工具** 为 Windows 可执行程序（`app.exe`）生成自签名证书、执行数字签名，以及在客户机实现无告警双击运行的完整工程实践。

---

## 📌 为什么需要数字签名？

1. **消除 SmartScreen 蓝底拦截警告**：
   Win10 / Win11 默认对未签名的程序弹出“*Windows 已保护你的电脑，未知的发布者*”拦截，用户必须点击“更多信息 -> 仍要运行”；
2. **极大降低杀毒软件误报（尤其结合 UPX 压缩）**：
   UPX 壳能将体积从 9.5MB 压至 3MB，但无签名的加壳程序极易被 360、火绒、Windows Defender 误报敏感。**一旦打上有效数字签名，杀软识别到明确签发者，误报率将大幅降低！**
3. **程序防伪与防篡改**：
   用户右键 `app.exe` -> 【属性】，将呈现专属的【数字签名】标签页，显示发布者名称与时间戳，文件一旦被病毒感染或篡改 1 字节签名立即失效。

---

## 🔑 核心概念：私钥与公钥是干嘛的？

* **私钥 (`.key` / `.pfx`) —— 锁在保险柜里的“实体公章”**：
  * 仅保存在开发者本地，**绝对严禁公开或提交到 GitHub**；
  * 用于对编译完成的 `app.exe` 进行密码学签名（盖章）。
* **公钥 (`.cer` / `.crt`) —— 公开给客户电脑的“印鉴认证卡”**：
  * 可以完全公开分发给客户；
  * **用途**：Windows 操作系统通过公钥验证软件确系你本人所发布、且未遭篡改；
  * **信任机制**：在客户电脑上将你的公钥导入一次【受信任的根证书颁发机构】后，该电脑就会永久信任你的公钥。此后你发布的任何版本 `app.exe` 双击直接秒开，永不弹警告！

---

## 🚀 完整实战操作流水线

### 第一步：生成专属代码签名证书 (0 外部依赖)

使用框架内置的纯 Go 标准库工具，跨平台一键生成 2048 位 RSA 代码签名根证书：

```bash
# 在项目根目录下执行 (参数名可自定义)：
go run github.com/xbt/godeniter/cmd/cert -name "我的软件开发工作室" -org "MyProject" -out ./certs
```

执行后将在 `./certs` 目录下生成：
* `app_codesign.key`：RSA 私钥（保密）
* `app_codesign.pfx`：Windows 标准 PKCS#12 签名包（默认密码 `123456`）
* `app_codesign.cer`：DER 格式公钥证书（用于分发给 Windows 客户机一键信任）
* `app_codesign.crt`：PEM 格式公钥文本

> ⚠️ **安全警告**：`./certs` 目录及 `*.key` / `*.pfx` 已被 `.gitignore` 自动拦截，**切勿强行提交到 GitHub 仓库**！

---

### 第二步：编译单文件程序

```bash
# macOS / Linux 下交叉编译 Windows 单文件
./build.sh

# 或 Windows 下执行
build.bat
```
产物输出在 `dist/app.exe`（自包含前端模板、应用图标与托盘，约 9.5MB）。

---

### 第三步：UPX 极速压缩 (可选，强烈推荐)

```bash
# 将 9.5MB 的 app.exe 无损压缩至 3MB 左右
upx --best dist/app.exe
```

---

### 第四步：给 `app.exe` 执行数字签名

> 🚨 **黄金铁律**：**必须严格按照【编译 -> UPX 压缩 -> 数字签名】的顺序执行！**
> 绝对严禁先签名后压缩，否则 UPX 会改写 PE 头导致数字签名被彻底破坏！

#### 方案 A：在 macOS / Linux 上签名 (免切 Windows 虚拟机)
通过 Homebrew 安装跨平台开源签名工具 `osslsigncode`：
```bash
brew install osslsigncode

# 执行代码签名并打上权威时间戳
osslsigncode sign \
  -pkcs12 certs/app_codesign.pfx \
  -pass 123456 \
  -n "Godeniter Application" \
  -ts "http://timestamp.digicert.com" \
  -in dist/app.exe \
  -out dist/app_signed.exe
```

#### 方案 B：在 Windows 上签名 (使用官方 signtool)
在 Windows 下打开终端（需装有 Windows SDK 或 Visual Studio）：
```cmd
signtool sign /f certs\app_codesign.pfx /p 123456 /tr http://timestamp.digicert.com /td sha256 /fd sha256 dist\app.exe
```

---

### 第五步：交付客户与客户机“一键信任”

将签名后的 `app_signed.exe`（或改名为 `app.exe`）和公钥证书 `app_codesign.cer` 一并打包交付给客户：

#### 客户机一键导入公钥（极速推荐）：
在 Windows 客户机上，以**管理员身份**打开 CMD 或 PowerShell，执行：
```cmd
certutil -addstore -f "ROOT" app_codesign.cer
```
*(提示：`CertUtil: -addstore 命令成功完成`)*

#### 图形界面手动导入：
1. 鼠标双击 `app_codesign.cer`；
2. 点击【安装证书】 -> 选择【本地计算机】 -> 下一步；
3. 选择【将所有的证书都放入下列存储】 -> 点击【浏览】；
4. 勾选并选择【受信任的根证书颁发机构】 -> 确定 -> 下一步完成。

---

## 🎯 最终交付效果

完成上述信任后，客户电脑上将获得顶级的桌面软件交付体验：
1. **直接双击 `app.exe`**：无黑框弹出、无 SmartScreen 蓝底危险拦截；
2. **托盘常驻**：屏幕右下角点亮精致图标，右键提供完整管理菜单；
3. **体积小巧**：仅 3MB 单文件，免装 Go 环境与运行时；
4. **右键属性**：显示公司/个人真实颁发者名称，权威时间戳防篡改！
