<?php

namespace App\Auth;

use Illuminate\Auth\GenericUser;

class MwUser extends GenericUser
{
    public function toArray(): array
    {
        return $this->attributes;
    }
}
