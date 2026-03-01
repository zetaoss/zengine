<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('runboxes', function (Blueprint $table) {
            $table->charset = 'utf8mb4';
            $table->collation = 'utf8mb4_general_ci';

            $table->increments('id');
            $table->string('hash', 255);
            $table->string('phase', 10);
            $table->unsignedInteger('user_id');
            $table->unsignedInteger('page_id');
            $table->string('type', 10);
            $table->json('payload');
            $table->json('outs')->nullable();
            $table->float('cpu')->nullable();
            $table->float('mem')->nullable();
            $table->float('time')->nullable();
            $table->timestamps();

            $table->unique('hash', 'unique_hash');
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('runboxes');
    }
};
