
############################################################
cat << EOF

maizenet

EOF
############################################################

apt update
apt install -y software-properties-common
add-apt-repository ppa:longsleep/golang-backports
apt install -y golang-go git
go get "github.com/gin-gonic/gin"
