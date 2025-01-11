<?php
namespace App\Http\Controllers;

use App\Jobs\CommonReportJob;
use App\Models\CommonReport;
use App\Models\CommonReportItem;
use Illuminate\Http\Request;

class CommonReportController extends MyController
{
    public function index()
    {
        return CommonReport::orderBy('id', 'desc')->paginate(15);
    }

    public function show(string $id): CommonReport
    {
        return CommonReport::findOrFail($id);
    }

    public function store(Request $req)
    {
        $err = $this->shouldCreatable();
        if ($err !== false) {
            return $err;
        }
        $names = array_filter(request('names', []));
        if (!is_array($names) || count($names) < 2) {
            return $this->newHTTPError(422, "비교 대상을 2개 이상 입력해 주세요.");
        }
        $report = new CommonReport;
        $report->user_id = $this->getMe()['avatar']['id'];
        $report->state = 0;
        $report->save();

        foreach ($names as $name) {
            $item = new CommonReportItem([
                'report_id' => $report->id,
                'name' => $name,
            ]);
            $item->save();
        }
        CommonReportJob::dispatch($report->id);
    }

    public function destroy($id)
    {
        $row = CommonReport::find($id);
        $err = $this->shouldDeletable($row->user_id);
        if ($err !== false) {
            return $err;
        }
        $row->delete();
        return ['status' => 'ok'];
    }
}
