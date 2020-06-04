
############################################################
cat << EOF

C . O . D . Y .

EOF
############################################################

apt update
apt install -y golang-go git
go get "github.com/gin-gonic/gin"
go get "github.com/gin-gonic/contrib/sessions"
go get "github.com/mattn/go-sqlite3"
go get "github.com/gorilla/websocket"
