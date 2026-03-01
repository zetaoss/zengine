<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('mw_statistics', function (Blueprint $table) {
            $table->id();
            $table->date('timeslot')->comment('KST date');
            $table->unsignedBigInteger('pages')->default(0);
            $table->unsignedBigInteger('articles')->default(0);
            $table->unsignedBigInteger('edits')->default(0);
            $table->unsignedBigInteger('images')->default(0);
            $table->unsignedBigInteger('users')->default(0);
            $table->unsignedBigInteger('activeusers')->default(0);
            $table->unsignedBigInteger('admins')->default(0);
            $table->unsignedBigInteger('jobs')->default(0);

            $table->unique('timeslot', 'mw_statistics_timeslot_unique');
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('mw_statistics');
    }
};
