<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        if (Schema::hasTable('stat_mw_hourly')) {
            return;
        }

        Schema::create('stat_mw_hourly', function (Blueprint $table) {
            $table->id();
            $table->dateTime('timeslot')->comment('UTC hour');
            $table->unsignedBigInteger('pages')->default(0);
            $table->unsignedBigInteger('articles')->default(0);
            $table->unsignedBigInteger('edits')->default(0);
            $table->unsignedBigInteger('images')->default(0);
            $table->unsignedBigInteger('users')->default(0);
            $table->unsignedBigInteger('activeusers')->default(0);
            $table->unsignedBigInteger('admins')->default(0);
            $table->unsignedBigInteger('jobs')->default(0);

            $table->unique('timeslot', 'stat_mw_hourly_timeslot_unique');
            $table->index('timeslot', 'stat_mw_hourly_timeslot_index');
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('stat_mw_hourly');
    }
};
