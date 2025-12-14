<?php

namespace App\Http\Controllers;

use App\Models\CommonReport;
use App\Services\CommonReportService;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Gate;

class CommonReportController extends Controller
{
    public function __construct(protected CommonReportService $service) {}

    public function index()
    {
        return CommonReport::orderByDesc('id')->paginate(15);
    }

    public function show(string $id)
    {
        return CommonReport::findOrFail($id);
    }

    public function store(Request $request)
    {
        Gate::authorize('unblocked');

        $names = $request->input('names', []);
        if (! is_array($names)) {
            return response()->json(['message' => '비교 대상을 2개 이상 입력해 주세요.'], 422);
        }

        $names = array_values(array_filter($names, fn ($v) => is_string($v) && trim($v) !== ''));
        if (count($names) < 2) {
            return response()->json(['message' => '비교 대상을 2개 이상 입력해 주세요.'], 422);
        }

        $userId = (int) auth()->id();

        return $this->service->create($names, $userId);
    }

    public function rerun(int $id)
    {
        $report = CommonReport::findOrFail($id);
        Gate::authorize('ownerOrSysop', (int) $report->user_id);

        $this->service->rerun($report);

        return ['ok' => true];
    }

    public function clone(int $id)
    {
        Gate::authorize('unblocked');
        $original = CommonReport::findOrFail($id);

        $userId = (int) auth()->id();

        return $this->service->clone($original, $userId);
    }

    public function destroy(int $id)
    {
        $report = CommonReport::find($id);
        if (! $report) {
            return response()->json(['message' => '해당 리포트가 없습니다.'], 404);
        }

        Gate::authorize('ownerOrSysop', (int) $report->user_id);
        $this->service->delete($report);

        return ['ok' => true];
    }
}
