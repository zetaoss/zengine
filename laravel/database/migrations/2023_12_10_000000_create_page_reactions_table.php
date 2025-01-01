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
        Schema::create('page_reactions', function (Blueprint $table) {
            $table->unsignedInteger('page_id');
            $table->unsignedInteger('cnt');
            $table->json('emoji_count');
            $table->primary('page_id');
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::dropIfExists('page_reactions');
    }
};
