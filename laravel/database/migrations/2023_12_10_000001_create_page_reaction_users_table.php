<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     */
    public function up(): void
    {
        Schema::create('page_reaction_users', function (Blueprint $table) {
            $table->unsignedInteger('page_id');
            $table->unsignedInteger('user_id');
            $table->unsignedInteger('emoji_code');
            $table->primary(['page_id', 'user_id', 'emoji_code']);
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('page_reaction_users');
    }
};
