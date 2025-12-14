<?php

namespace App\Providers;

use Illuminate\Foundation\Support\Providers\AuthServiceProvider as ServiceProvider;
use Illuminate\Support\Facades\Gate;

class AuthServiceProvider extends ServiceProvider
{
    public function boot(): void
    {
        $this->registerPolicies();

        Gate::define('unblocked', function ($user) {
            if (! $user) {
                return false;
            }

            return ! isset($user->blockid)
                && ! isset($user->blockedbyid)
                && ! isset($user->blockedby)
                && ! isset($user->blockexpiry);
        });

        Gate::define('owner', function ($user, int $ownerId) {
            if (! $user) {
                return false;
            }

            return (int) $user->id === (int) $ownerId;
        });

        Gate::define('ownerOrSysop', function ($user, int $ownerId) {
            if (! $user) {
                return false;
            }

            if ((int) $user->id === (int) $ownerId) {
                return true;
            }

            return in_array('sysop', (array) ($user->groups ?? []), true);
        });
    }
}
