#!/usr/bin/env bash

# Switch to root before executing this script
# Do **NOT** using `sudo` with this script

set -e

SD_CARD=$1
INDEX=$2
ROOTFS_TAR="ArchLinuxARM-rpi-aarch64-latest.tar.gz"
MY_HOME="/home/young"
PI_HOST_NAME="$(basename $MY_HOME)-pi$INDEX"

# see https://archlinuxarm.org/platforms/armv8/broadcom/raspberry-pi-4
# wget http://os.archlinuxarm.org/os/ArchLinuxARM-rpi-aarch64-latest.tar.gz

if [[ ! -b $SD_CARD ]]; then
    echo "Usage: $0 <path/to/sdcard> <pi-num>"
    exit 1
fi

if [[ $USER != "root"  ]]; then
    echo "Current user is $USER, Please switch to root before executing this script"
    exit 1
fi

function part_sdcard() {
    # see https://wiki.archlinux.org/title/Parted
    # The mkpart command does not actually create the file system:
    # the fs-type parameter will simply be used by parted to set a 1-byte code
    # that is used by boot loaders to "preview" what kind of data
    # is found in the partition, and act accordingly if necessary.

    echo "Create new partition table on $SD_CARD"
    parted --script $SD_CARD -- mklabel msdos

    echo "Create partition for boot"
    parted --script $SD_CARD -- mkpart primary fat32 1MiB 200MiB set 1 boot on

    echo "Create partition for rootfs"
    parted --script $SD_CARD -- mkpart primary ext4 200MiB -10MiB

    partprobe $SD_CARD
    parted $SD_CARD unit GiB print

    echo "mkfs.vfat on ${SD_CARD}1"
    mkfs.vfat ${SD_CARD}1

    echo "mkfs.ext4 on ${SD_CARD}2"
    mkfs.ext4 -F ${SD_CARD}2

    sleep 3 && sync
}

function write_image(){
    echo "Mount ${SD_CARD}1 at boot/"
    mkdir -p boot
    mount ${SD_CARD}1 boot

    echo "Mount ${SD_CARD}2 at rootfs/"
    mkdir -p rootfs
    mount ${SD_CARD}2 rootfs

    echo "Extract ${ROOTFS_TAR} to rootfs/ ..."
    bsdtar -xpf ${ROOTFS_TAR} -C rootfs
    echo "Copy boot files to boot/ ..."
    mv rootfs/boot/* boot

    echo "Sync content to ${SD_CARD} ..."
    sleep 3 && sync
}

function add_preconfig() {
    echo "Add extra tools to rootfs ..."

    cp ./config_pi.sh rootfs/root/
    cp -rf ${MY_HOME}/.ssh rootfs/root/

    cp -RPf extra/multi-user.target.wants rootfs/etc/systemd/system/
    cp -f extra/locale.gen rootfs/etc/locale.gen
    cp -f extra/nsswitch.conf rootfs/etc/nsswitch.conf
    cp -f extra/proxychains.conf rootfs/etc/proxychains.conf
    cp -f extra/wlan0.network rootfs/etc/systemd/network/
    cp -f extra/mirrorlist rootfs/etc/pacman.d/mirrorlist

    echo "Set hostname to $PI_HOST_NAME"
    # no systemd daemon lives in this env, following cmd won't work:
    # hostnamectl set-hostname --transient --static --pretty $PI_HOST_NAME
    printf "%s\n" $PI_HOST_NAME > rootfs/etc/hostname

    chown -R root:root rootfs/root/
    chown -R root:root rootfs/etc/systemd/system/
    chown -R root:root rootfs/etc/hostname
    chown -R root:root rootfs/etc/locale.gen
    chown -R root:root rootfs/etc/nsswitch.conf
    chown -R root:root rootfs/etc/proxychains.conf
    chown -R root:root rootfs/etc/systemd/network/
    chown -R root:root rootfs/etc/pacman.d/mirrorlist

    echo "Sync content to ${SD_CARD} ..."
    sleep 3 && sync
}


function chroot_setup() {
    # see https://disconnected.systems/blog/raspberry-pi-archlinuxarm-setup/
    # see https://raspberrypi.stackexchange.com/questions/855/is-it-possible-to-update-upgrade-and-install-software-before-flashing-an-image
    # see https://wiki.archlinux.org/title/QEMU#Chrooting_into_arm/arm64_environment_from_x86_64
    # see https://nerdstuff.org/posts/2020/2020-003_simplest_way_to_create_an_arm_chroot/

    # yay -S binfmt-qemu-static qemu-user-static arch-install-scripts
    # systemctl restart systemd-binfmt.service

    echo "Chroot to ARM rootfs/"
    mkdir -p rootfs/secret
    mount -o ro,bind secret/ rootfs/secret
    mount ${SD_CARD}1 rootfs/boot

    # sleep infinity  # for debug

    echo "Config pi evn ..."
    arch-chroot rootfs /root/config_pi.sh

    echo "Sync content to ${SD_CARD} ..."
    sleep 3 && sync

    echo "Umount boot and rootfs"
    umount -l rootfs/boot rootfs/secret
    rm -r rootfs/secret
    umount -l boot rootfs
    rm -r boot rootfs
}


function main() {
    echo "Setup Archlinux ARM on $SD_CARD"
    part_sdcard
    write_image
    add_preconfig
    chroot_setup
    echo "Setup Archlinux ARM to ${SD_CARD} finished."
}

main
