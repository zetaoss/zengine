<?php

namespace App\Http\Controllers;

use App\Services\UserService;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\DB;
use Laravel\Socialite\Facades\Socialite;

class SocialController extends Controller
{
    public function redirect(Request $request, string $provider)
    {
        $returnto = $request->input('returnto', '');
        if (! $returnto || strlen($returnto) < 1 || strlen($returnto) > 255) {
            $returnto = '';
        }
        $request->session()->put('returnto', $returnto);

        return Socialite::driver($provider)->redirect();
    }

    public function callback(Request $request, string $provider)
    {
        $user = Socialite::driver($provider)->user();
        $social_id = $user->getId();
        $mwdb = DB::connection('mwdb');
        $row = $mwdb->table('user_social')->select('id', 'user_id')->where([
            ['provider', '=', $provider],
            ['social_id', '=', $social_id],
        ])->first();
        if (! $row) {
            // first social
            $id = $mwdb->table('user_social')->insertGetId(['provider' => $provider, 'social_id' => $social_id]);
            $row = $mwdb->table('user_social')->where('id', $id)->first();
        }
        $user_id = $row->user_id;
        if (! is_numeric($user_id) || $user_id < 1) {
            // need join
            $code = sha1(rand());
            Cache::store('redis')->put("code:$code", $row->id, 9);

            return redirect("/social-join/$code");
        }
        // login
        $path = $this->getBridgePath($user_id);
        $returnto = session('returnto', false);
        if (! $returnto) {
            return redirect($path);
        }

        return redirect("$path&returnto=$returnto");
    }

    public function checkCode($code)
    {
        return ['status' => 'success', 'data' => $this->validateCode($code)];
    }

    public function loginCode($code)
    {
        if (! $this->validateCode($code)) {
            return $this->newHTTPError(403, 'invalid code');
        }
        $id = Cache::store('redis')->get("code:$code");
        Cache::store('redis')->forget("code:$code");
        $row = DB::connection('mwdb')->table('user_social')->select('user_id')->where('id', $id)->first();
        if (! $row || ! is_numeric($row->user_id) || $row->user_id < 1) {
            return $this->newHTTPError(403, 'invalid user id');
        }
        $path = $this->getBridgePath($row->user_id);

        return ['status' => 'success', 'data' => $path];
    }

    private function validateCode($code)
    {
        $id = Cache::store('redis')->get("code:$code");
        if (! $id || ! is_numeric($id) || $id < 1) {
            return false;
        }

        return true;
    }

    private function getBridgePath($user_id)
    {
        $otp = sha1(rand());
        Cache::store('redis')->put("otp:$otp", $user_id, 30);
        $username = UserService::getUserAvatar($user_id)['name'] ?? '';

        return "/social-bridge?otp=$otp&username=$username";
    }
}
