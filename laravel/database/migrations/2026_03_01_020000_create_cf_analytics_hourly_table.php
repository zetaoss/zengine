<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('cf_analytics_hourly', function (Blueprint $table) {
            $table->id();
            $table->dateTime('timeslot')->comment('UTC');
            $table->string('name', 64);
            $table->text('value')->comment('JSON text');

            $table->unique(['timeslot', 'name'], 'cf_analytics_hourly_timeslot_name_unique');
            $table->index('timeslot', 'cf_analytics_hourly_timeslot_index');
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('cf_analytics_hourly');
    }
};
