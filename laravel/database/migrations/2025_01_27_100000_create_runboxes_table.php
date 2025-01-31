<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('runboxes', function (Blueprint $table) {
            $table->id();
            $table->string('hash', 255)->unique();
            $table->tinyInteger('step');
            $table->unsignedInteger('user_id');
            $table->unsignedInteger('page_id');
            $table->string('type', 8);
            $table->json('payload');
            $table->json('outs')->nullable();
            $table->unsignedInteger('cpu')->default(0);
            $table->unsignedInteger('mem')->default(0);
            $table->unsignedInteger('time')->default(0);
            $table->timestamps();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('runboxes');
    }
};
