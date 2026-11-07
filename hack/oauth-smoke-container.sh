set -eu
cd /var/www/html
mkdir -p /tmp/mwdb
php maintenance/install.php --dbtype sqlite --dbpath /tmp/mwdb --server http://127.0.0.1:8080 --scriptpath /w --lang en --pass TemporaryPass123 'Test Wiki' TestAdmin >/tmp/install.log
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out /tmp/oauth-private.pem 2>/dev/null
openssl rsa -in /tmp/oauth-private.pem -pubout -out /tmp/oauth-public.pem 2>/dev/null
export MW_OAUTH2_PRIVATE_KEY="$(cat /tmp/oauth-private.pem)"
export MW_OAUTH2_PUBLIC_KEY="$(cat /tmp/oauth-public.pem)"
cat >> LocalSettings.php <<'PHP'
wfLoadExtension('OAuth');
$wgOAuth2PrivateKey = getenv('MW_OAUTH2_PRIVATE_KEY');
$wgOAuth2PublicKey = getenv('MW_OAUTH2_PUBLIC_KEY');
$wgGroupPermissions['user']['mwoauthproposeconsumer'] = true;
$wgGroupPermissions['user']['mwoauthupdateownconsumer'] = true;
PHP
php maintenance/run.php update --quick >/tmp/update.log
printf '%s\n' '$u = User::newFromName("TestAdmin"); $u->setEmail("test@example.com"); $u->setEmailAuthenticationTimestamp(wfTimestampNow()); $u->saveSettings();' | php maintenance/run.php eval.php >/tmp/eval.log
ln -s /var/www/html /var/www/html/w
php -S 127.0.0.1:8080 -t /var/www/html >/tmp/server.log 2>&1 &
python3 /tmp/mw-oauth-flow.py
