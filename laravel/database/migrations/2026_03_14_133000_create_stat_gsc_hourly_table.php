<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        if (Schema::hasTable('stat_gsc_hourly')) {
            return;
        }

        Schema::create('stat_gsc_hourly', function (Blueprint $table) {
            $table->id();
            $table->dateTime('timeslot')->comment('UTC hour derived from Search Console Pacific time');
            $table->unsignedBigInteger('clicks')->default(0);
            $table->unsignedBigInteger('impressions')->default(0);
            $table->decimal('ctr', 9, 4)->default(0);
            $table->decimal('position', 9, 4)->default(0);

            $table->unique('timeslot', 'stat_gsc_hourly_timeslot_unique');
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('stat_gsc_hourly');
    }
};
