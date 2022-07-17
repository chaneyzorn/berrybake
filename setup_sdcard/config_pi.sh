#!/usr/bin/env bash

NEW_USER=young

echo "Remove alarm user"
userdel -r alarm

echo "Create and config user for $NEW_USER"
# see https://unix.stackexchange.com/questions/81240/manually-generate-password-for-etc-shadow
# openssl passwd -6 -salt $salt $yourpass
# -1 MD5
# -5 SHA256
# -6 SHA512
useradd -p $(cat /secret/passwd.sha512) -m $NEW_USER
gpasswd -a $NEW_USER wheel

cp -r /root/.ssh /home/$NEW_USER/
cp /home/$NEW_USER/.ssh/id_*.pub /home/$NEW_USER/.ssh/authorized_keys
chown -R $NEW_USER:$NEW_USER /home/$NEW_USER/.ssh

# echo "Setup clash config"
# mkdir -p /home/$NEW_USER/.config
# cp -r /secret/clash /home/$NEW_USER/.config/
# chown -R $NEW_USER:$NEW_USER /home/$NEW_USER/.config

echo "Setup wpa_supplicant for wlan0"
# see https://wiki.archlinux.org/title/Wpa_supplicant#Connecting_with_wpa_passphrase
# wpa_passphrase <MYSSID> <passphrase>
cp -f  /secret/wpa_supplicant-wlan0.conf /etc/wpa_supplicant/
chown -R root:root /etc/wpa_supplicant/

echo "Set locales"
locale-gen

cat > /etc/locale.conf << EOF
LANG=en_US.UTF-8
LC_ADDRESS=en_US.UTF-8
LC_IDENTIFICATION=en_US.UTF-8
LC_MEASUREMENT=en_US.UTF-8
LC_MONETARY=en_US.UTF-8
LC_NAME=en_US.UTF-8
LC_NUMERIC=en_US.UTF-8
LC_PAPER=en_US.UTF-8
LC_TELEPHONE=en_US.UTF-8
LC_TIME=en_US.UTF-8
EOF

echo "Set Time Zone to Asia/Shanghai"
ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

pacman-key --init
pacman-key --populate archlinuxarm

# pacman -Syu --noconfirm
pacman -Sy
pacman -S --noconfirm base-devel wget curl git man-db \
clash proxychains sudo vim neovim rsync htop avahi nss-mdns docker prometheus-node-exporter

echo "Set sudoers for $NEW_USER"
cat > /etc/sudoers.d/$NEW_USER << EOF
$NEW_USER ALL=(ALL) ALL
EOF

# ALL_PROXY="socks5://127.0.0.1:1080" yay -S yay-bin --overwrite /usr/bin/yay
