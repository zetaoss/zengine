set -Eeuo pipefail
trap 'for log in /tmp/install.log /tmp/update.log /tmp/eval.log /tmp/server.log; do if [[ -f "$log" ]]; then echo "--- $log ---" >&2; tail -n 50 "$log" >&2; fi; done' ERR
cd /var/www/html
mkdir -p /tmp/mwdb
env -u MW_OAUTH2_PRIVATE_KEY -u MW_OAUTH2_PUBLIC_KEY php -r '
function wfLoadExtension($name) { global $loaded; $loaded[] = $name; }
require "/tmp/BaseExtensionSettings.php";
if (!in_array("OAuth", $loaded, true) || $wgOAuth2PrivateKey !== "" || $wgOAuth2PublicKey !== "") { exit(1); }
'
php maintenance/install.php --dbtype sqlite --dbpath /tmp/mwdb --server http://127.0.0.1:8080 --scriptpath /w --lang en --pass TemporaryPass123 'Test Wiki' TestAdmin >/tmp/install.log
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out /tmp/oauth-private.pem 2>/dev/null
openssl rsa -in /tmp/oauth-private.pem -pubout -out /tmp/oauth-public.pem 2>/dev/null
export MW_OAUTH2_PRIVATE_KEY="$(cat /tmp/oauth-private.pem)"
export MW_OAUTH2_PUBLIC_KEY="$(cat /tmp/oauth-public.pem)"
# The base image contains external extensions; the production image adds ZetaExtension later.
mkdir -p extensions/ZetaExtension
printf '%s\n' '{"name":"ZetaExtension","version":"test","manifest_version":2}' > extensions/ZetaExtension/extension.json
cat >> LocalSettings.php <<'PHP'
require '/tmp/BaseExtensionSettings.php';
PHP
php maintenance/run.php update --quick >/tmp/update.log
printf '%s\n' '$u = User::newFromName("TestAdmin"); $u->setEmail("test@example.com"); $u->setEmailAuthenticationTimestamp(wfTimestampNow()); $u->saveSettings();' | php maintenance/run.php eval.php >/tmp/eval.log
ln -s /var/www/html /var/www/html/w
php -S 127.0.0.1:8080 -t /var/www/html >/tmp/server.log 2>&1 &
python3 /tmp/mw-oauth-flow.py
