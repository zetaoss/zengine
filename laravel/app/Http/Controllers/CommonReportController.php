<?php

namespace App\Http\Controllers;

use App\Models\CommonReport;
use App\Services\CommonReportService;
use Illuminate\Http\Request;

class CommonReportController extends MyController
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
        if ($err = $this->shouldCreatable()) {
            return $err;
        }

        $names = array_filter($request->input('names', []));
        if (! is_array($names) || count($names) < 2) {
            return $this->newHTTPError(422, '비교 대상을 2개 이상 입력해 주세요.');
        }

        $userId = $this->getMe()['avatar']['id'];

        return $this->service->create($names, $userId);
    }

    public function rerun(int $id)
    {
        $report = CommonReport::findOrFail($id);
        if ($err = $this->shouldDeletable($report->user_id)) {
            return $err;
        }

        $this->service->rerun($report);
    }

    public function clone(int $id)
    {
        $original = CommonReport::findOrFail($id);
        if ($err = $this->shouldCreatable()) {
            return $err;
        }

        $userId = $this->getMe()['avatar']['id'];

        return $this->service->clone($original, $userId);
    }

    public function destroy($id)
    {
        $report = CommonReport::find($id);
        if (! $report) {
            return $this->newHTTPError(404, '해당 리포트가 없습니다.');
        }

        if ($err = $this->shouldDeletable($report->user_id)) {
            return $err;
        }

        $this->service->delete($report);

        return ['status' => 'ok'];
    }
}
