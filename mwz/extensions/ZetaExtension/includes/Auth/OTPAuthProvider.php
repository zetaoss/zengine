<?php
namespace ZetaExtension\Auth;

use MediaWiki\Auth\AbstractPrimaryAuthenticationProvider;
use MediaWiki\Auth\AuthenticationRequest;
use MediaWiki\Auth\AuthenticationResponse;
use User;

class OTPAuthProvider extends AbstractPrimaryAuthenticationProvider
{
    public function beginPrimaryAuthentication($reqs)
    {
        $req = $reqs[0] ?? false;
        if (!$req) {
            return AuthenticationResponse::newAbstain();
        }
        $redis = new \Redis();
        $redis->connect(getenv('REDIS_HOST'));
        $userID = $redis->get("otp:{$req->password}");
        $user = User::newFromId($userID);
        if (!$user || $req->username != $user->getName()) {
            return AuthenticationResponse::newAbstain();
        }
        return AuthenticationResponse::newPass($user->getName());
    }
    public function continuePrimaryAuthentication($reqs)
    {
        return null;
    }
    public function testUserCanAuthenticate($username)
    {
        return null;
    }
    public function testUserExists($username, $flags = User::READ_NORMAL)
    {
        return false;
    }
    public function providerAllowsPropertyChange($property)
    {
        return false;
    }
    public function providerAllowsAuthenticationDataChange(AuthenticationRequest $req, $checkData = true)
    {
        return \StatusValue::newGood('ignored');
    }
    public function providerChangeAuthenticationData(AuthenticationRequest $req)
    {
        return null;
    }
    public function accountCreationType()
    {
        return self::TYPE_CREATE;
    }
    public function beginPrimaryAccountCreation($user, $creator, $reqs)
    {
        return AuthenticationResponse::newAbstain();
    }
    public function getAuthenticationRequests($action, $options)
    {
        return [];
    }
}
