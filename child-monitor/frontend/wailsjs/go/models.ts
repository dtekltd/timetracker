export namespace models {
	
	export class ActivitySample {
	    id: number;
	    // Go type: time
	    sampled_at: any;
	    process_name: string;
	    process_path: string;
	    window_title: string;
	    window_handle: string;
	    is_idle: boolean;
	    idle_seconds: number;
	
	    static createFrom(source: any = {}) {
	        return new ActivitySample(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sampled_at = this.convertValues(source["sampled_at"], null);
	        this.process_name = source["process_name"];
	        this.process_path = source["process_path"];
	        this.window_title = source["window_title"];
	        this.window_handle = source["window_handle"];
	        this.is_idle = source["is_idle"];
	        this.idle_seconds = source["idle_seconds"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AppStatus {
	    version: string;
	    monitoring_paused: boolean;
	    auto_start_enabled: boolean;
	    screenshot_folder: string;
	
	    static createFrom(source: any = {}) {
	        return new AppStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.monitoring_paused = source["monitoring_paused"];
	        this.auto_start_enabled = source["auto_start_enabled"];
	        this.screenshot_folder = source["screenshot_folder"];
	    }
	}
	export class DailyAppUsage {
	    usage_date: string;
	    process_name: string;
	    app_name: string;
	    total_seconds: number;
	    active_seconds: number;
	    idle_seconds: number;
	    open_count: number;
	    last_window_title: string;
	
	    static createFrom(source: any = {}) {
	        return new DailyAppUsage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.usage_date = source["usage_date"];
	        this.process_name = source["process_name"];
	        this.app_name = source["app_name"];
	        this.total_seconds = source["total_seconds"];
	        this.active_seconds = source["active_seconds"];
	        this.idle_seconds = source["idle_seconds"];
	        this.open_count = source["open_count"];
	        this.last_window_title = source["last_window_title"];
	    }
	}
	export class DailySummary {
	    usage_date: string;
	    total_seconds: number;
	    active_seconds: number;
	    idle_seconds: number;
	    screenshot_count: number;
	    top_app: string;
	
	    static createFrom(source: any = {}) {
	        return new DailySummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.usage_date = source["usage_date"];
	        this.total_seconds = source["total_seconds"];
	        this.active_seconds = source["active_seconds"];
	        this.idle_seconds = source["idle_seconds"];
	        this.screenshot_count = source["screenshot_count"];
	        this.top_app = source["top_app"];
	    }
	}
	export class Screenshot {
	    id: number;
	    // Go type: time
	    captured_at: any;
	    file_path: string;
	    file_name: string;
	    file_size: number;
	    width: number;
	    height: number;
	    display_index: number;
	    upload_status: string;
	
	    static createFrom(source: any = {}) {
	        return new Screenshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.captured_at = this.convertValues(source["captured_at"], null);
	        this.file_path = source["file_path"];
	        this.file_name = source["file_name"];
	        this.file_size = source["file_size"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.display_index = source["display_index"];
	        this.upload_status = source["upload_status"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DashboardData {
	    date: string;
	    summary: DailySummary;
	    top_apps: DailyAppUsage[];
	    latest_screenshots: Screenshot[];
	
	    static createFrom(source: any = {}) {
	        return new DashboardData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.summary = this.convertValues(source["summary"], DailySummary);
	        this.top_apps = this.convertValues(source["top_apps"], DailyAppUsage);
	        this.latest_screenshots = this.convertValues(source["latest_screenshots"], Screenshot);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

