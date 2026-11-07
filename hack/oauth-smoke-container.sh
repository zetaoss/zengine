set -eu
cd /var/www/html
mkdir -p /tmp/mwdb
php maintenance/install.php --dbtype sqlite --dbpath /tmp/mwdb --server http://127.0.0.1:8080 --scriptpath /w --lang en --pass TemporaryPass123 'Test Wiki' TestAdmin >/tmp/install.log
mkdir -p /run/secrets/mediawiki-oauth
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out /run/secrets/mediawiki-oauth/private.pem 2>/dev/null
openssl rsa -in /run/secrets/mediawiki-oauth/private.pem -pubout -out /run/secrets/mediawiki-oauth/public.pem 2>/dev/null
cat >> LocalSettings.php <<'PHP'
wfLoadExtension('OAuth');
$wgOAuth2PrivateKey = 'file:///run/secrets/mediawiki-oauth/private.pem';
$wgOAuth2PublicKey = 'file:///run/secrets/mediawiki-oauth/public.pem';
$wgGroupPermissions['user']['mwoauthproposeconsumer'] = true;
$wgGroupPermissions['user']['mwoauthupdateownconsumer'] = true;
PHP
php maintenance/run.php update --quick >/tmp/update.log
printf '%s\n' '$u = User::newFromName("TestAdmin"); $u->setEmail("test@example.com"); $u->setEmailAuthenticationTimestamp(wfTimestampNow()); $u->saveSettings();' | php maintenance/run.php eval.php >/tmp/eval.log
ln -s /var/www/html /var/www/html/w
php -S 127.0.0.1:8080 -t /var/www/html >/tmp/server.log 2>&1 &
python3 /tmp/mw-oauth-flow.py
