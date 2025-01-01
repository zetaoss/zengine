<?php
namespace App\Http\Controllers;

use App\Models\Oneline;
use Illuminate\Http\Request;

class OnelineController extends Controller
{
    public function recent()
    {
        return Oneline::orderBy('id', 'desc')->take(15)->get();
    }

    public function index()
    {
        //
    }

    public function show(Oneline $oneline)
    {
        //
    }

    public function store(Request $request)
    {
        //
    }

    public function update(Request $request, Oneline $oneline)
    {
        //
    }

    public function destroy(Oneline $oneline)
    {
        //
    }
}
