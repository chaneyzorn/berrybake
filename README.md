# Berry Bake —— 树莓派集群环境

## 树莓派安装 archlinux

树莓派安装 archlinux 系统参考了以下资料：

- [archlinuxarm 官网](https://archlinuxarm.org/platforms/armv8/broadcom/raspberry-pi-4)
- [Parted - archlinux wiki](https://wiki.archlinux.org/title/Parted) parted 在分区时，并没有真正的创建文件系统，而是只写入了一字节的标记，用于引导程序预览分区可能的文件系统。

使用 x86 环境制作树莓派 arm 根文件系统参考了以下资料，其中的关键是使用 `qemu-user-static`：

- [Create Custom ArchlinuxArm Images for the Raspberry Pi](https://disconnected.systems/blog/raspberry-pi-archlinuxarm-setup/)
- [Is it possible to update, upgrade and install software before flashing an image?](https://raspberrypi.stackexchange.com/questions/855/is-it-possible-to-update-upgrade-and-install-software-before-flashing-an-image)
- [QEMU-Chrooting_into_arm](https://wiki.archlinux.org/title/QEMU#Chrooting_into_arm/arm64_environment_from_x86_64)
- [The Simplest Way To Create an ARM Chroot in Arch Linux](https://nerdstuff.org/posts/2020/2020-003_simplest_way_to_create_an_arm_chroot/)

预定制树莓派环境参考了以下资料：

- [Manually generate password for /etc/shadow](https://unix.stackexchange.com/questions/81240/manually-generate-password-for-etc-shadow) 预配置帐号密码
- [Wireless#WPA2_Personal](https://wiki.archlinux.org/index.php/Network_configuration/Wireless#WPA2_Personal) WPA2 认证需要使用 wpa_supplicant
- [wpa_supplicant](https://wiki.archlinux.org/index.php/Wpa_supplicant)
- [Wpa_supplicant-Connecting_with_wpa_passphrase](https://wiki.archlinux.org/title/Wpa_supplicant#Connecting_with_wpa_passphrase) 预配置 wifi 密码
- [Avahi](https://wiki.archlinux.org/title/avahi) 零配置网络，提供本地域名解析服务，直接通过 hostname 即可访问树莓派，无需预先知道被分配给树莓派的 IP 地址。

另外，chroot 环境中没有实际对应的 systemd 服务，因此只能手工生成启动服务的链接，无法使用 `systemctl enable <service>` 命令。

BCM2711 Stepping C0 无法使用 aarch64 镜像中的 uboot 正常启动,使用 `pacman -S linux-rpi` 安装上游内核后解决：

- ["error -5 whilst initializing SD card"@20220315 aarch64 img](https://archlinuxarm.org/forum/viewtopic.php?f=65&t=15994#p69312)
- [raspberry-pi-4-model-bs-arriving-newer-c0-stepping](https://www.jeffgeerling.com/blog/2021/raspberry-pi-4-model-bs-arriving-newer-c0-stepping)
- [Raspberry Pi 4 Model B 8 GB (2/3 Fail with SD Card error)](https://archlinuxarm.org/forum/viewtopic.php?f=67&t=15422&start=20#p67299)

## rpifand 风扇控制守护进程

rpifand 实现 GPIO PWM 风扇控制参考了以下资料：

- [Argon mini FAN](https://item.taobao.com/item.htm?id=634516381454)
- [pinout.xyz](https://pinout.xyz/) Pinout! The Raspberry Pi GPIO pinout guide.
- [Raspberry pi 4 GPIO controlled cooling fan](https://www.hackster.io/talofer99/raspberry-pi-4-gpio-controlled-cooling-fan-20fe85) 初期最重要的一篇文章，让我初次实现了真实的风扇起停控制，给了我很大的信心。
- [gpiozero](https://github.com/gpiozero/gpiozero/issues/707) gpiozero 在程序退出时会执行固定的清理动作，这虽然对初学者比较友好，但是却失去了定制特殊行为的灵活性。这就是 rpifand 不使用 gpiozero 的原因。
- [go-rpio](https://github.com/stianeikeland/go-rpio) 一个使用 golang 编写的用于树莓派的 GPIO 库，简洁易用，注释清晰，品质可靠。

rpifand 的 archlinux 软件包构建参考了以下资料：

- [创建 archlinux 下的软件包](https://wiki.archlinux.org/title/Creating_packages)
- [PKGBUILD 中的常用定义量](https://wiki.archlinux.org/title/PKGBUILD)
- [archlinuxcn 为 nerdctl 编写的 PKGBUILD](https://github.com/archlinuxcn/repo/blob/master/archlinuxcn/nerdctl/PKGBUILD)
- [go/issues/44542](https://github.com/golang/go/issues/44542) namcap 抱怨 “lacks FULL RELRO, check LDFLAGS”，虽然不知道具体原因，但是参考这里解决了问题
- [go cmd/link: support full relro](https://groups.google.com/g/golang-codereviews/c/b9mupWkiYDk)

## TODO

- [x] 将树莓派镜像写入 SD 卡的过程脚本化，并提供定制能力
- [x] 预配置 ssh 帐号密码
- [x] 预配置 wpa_supplicant wifi 密码认证，开机自动连接
- [x] 预配置 Avahi 零配置网络
- [x] 使用 prometheus 和 node-exporter 监控树莓派集群
- [x] 树莓派 GPIO 风扇守护进程解决散热问题
- [x] 风扇守护进程 archlinux 软件包构建
- [x] 风扇守护进程的信号处理，在程序退出时，将风扇置于长运行状态
- [ ] 一套简单的 web 管理界面
- [ ] 集群化，比如引入 k8s，引入对象存储

## license

MIT
