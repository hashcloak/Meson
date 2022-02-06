# /bin/ash

echo ""
echo "Update Voting TrustOptions..."
echo ""
/go/bin/updateconfig -f /conf/katzenpost.toml

echo ""
echo "Start Meson Server"
echo ""
/go/bin/server -f /conf/katzenpost.toml