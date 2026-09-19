export namespace ai {
	
	export class ModelInfo {
	    ID: string;
	    Name: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	    }
	}

}

export namespace languages {
	
	export class Language {
	    Name: string;
	    Code: string;
	
	    static createFrom(source: any = {}) {
	        return new Language(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Code = source["Code"];
	    }
	}

}

export namespace main {
	
	export class AICompletionRequest {
	    Task: string;
	    Text: string;
	    Command: string;
	    Provider: string;
	
	    static createFrom(source: any = {}) {
	        return new AICompletionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Task = source["Task"];
	        this.Text = source["Text"];
	        this.Command = source["Command"];
	        this.Provider = source["Provider"];
	    }
	}
	export class AIProvider {
	    Name: string;
	    Kind: string;
	    BaseURL: string;
	    Model: string;
	    Temperature: number;
	    NoAPIKey: boolean;
	    LowReasoning: boolean;
	    Custom: boolean;
	    HasAPIKey: boolean;
	    Configured: boolean;
	    Active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AIProvider(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Kind = source["Kind"];
	        this.BaseURL = source["BaseURL"];
	        this.Model = source["Model"];
	        this.Temperature = source["Temperature"];
	        this.NoAPIKey = source["NoAPIKey"];
	        this.LowReasoning = source["LowReasoning"];
	        this.Custom = source["Custom"];
	        this.HasAPIKey = source["HasAPIKey"];
	        this.Configured = source["Configured"];
	        this.Active = source["Active"];
	    }
	}
	export class MachineInfo {
	    Cores: number;
	    MemoryMB: number;
	
	    static createFrom(source: any = {}) {
	        return new MachineInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Cores = source["Cores"];
	        this.MemoryMB = source["MemoryMB"];
	    }
	}
	export class SpeechModel {
	    ID: string;
	    Name: string;
	    Description: string;
	    SizeMB: number;
	    Languages: languages.Language[];
	    LanguageDetect: boolean;
	    Streaming: boolean;
	    Custom: boolean;
	    Accuracy: number;
	    Speed: number;
	    Fit: string;
	    FitLabel: string;
	    Downloaded: boolean;
	    Downloading: boolean;
	    Recommended: boolean;
	    Suggested: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SpeechModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Description = source["Description"];
	        this.SizeMB = source["SizeMB"];
	        this.Languages = this.convertValues(source["Languages"], languages.Language);
	        this.LanguageDetect = source["LanguageDetect"];
	        this.Streaming = source["Streaming"];
	        this.Custom = source["Custom"];
	        this.Accuracy = source["Accuracy"];
	        this.Speed = source["Speed"];
	        this.Fit = source["Fit"];
	        this.FitLabel = source["FitLabel"];
	        this.Downloaded = source["Downloaded"];
	        this.Downloading = source["Downloading"];
	        this.Recommended = source["Recommended"];
	        this.Suggested = source["Suggested"];
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
	export class SpeechService {
	    ID: string;
	    Name: string;
	    BaseURL: string;
	    Models: string[];
	    KeyHint: string;
	    Custom: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SpeechService(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.BaseURL = source["BaseURL"];
	        this.Models = source["Models"];
	        this.KeyHint = source["KeyHint"];
	        this.Custom = source["Custom"];
	    }
	}

}

export namespace notion {
	
	export class NotionBlock {
	    Block: any;
	    Children: NotionBlock[];
	
	    static createFrom(source: any = {}) {
	        return new NotionBlock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Block = source["Block"];
	        this.Children = this.convertValues(source["Children"], NotionBlock);
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

export namespace repository {
	
	export class Cache {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    key: string;
	    // Go type: datatypes
	    value: any;
	
	    static createFrom(source: any = {}) {
	        return new Cache(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.key = source["key"];
	        this.value = this.convertValues(source["value"], null);
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
	export class CacheValue {
	    value: any;
	
	    static createFrom(source: any = {}) {
	        return new CacheValue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	    }
	}
	export class Config {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    TranscriberSource: string;
	    SpeechModelID?: string;
	    RemoteProvider: string;
	    RemoteModel: string;
	    RemoteBaseURL: string;
	    AIProvider: string;
	    AIProviders: number[];
	    NotionApiKey?: string;
	    OpenAIApiKey?: string;
	    GroqApiKey?: string;
	    LocalWhisperModel?: string;
	    LocalWhisperGPU?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.TranscriberSource = source["TranscriberSource"];
	        this.SpeechModelID = source["SpeechModelID"];
	        this.RemoteProvider = source["RemoteProvider"];
	        this.RemoteModel = source["RemoteModel"];
	        this.RemoteBaseURL = source["RemoteBaseURL"];
	        this.AIProvider = source["AIProvider"];
	        this.AIProviders = source["AIProviders"];
	        this.NotionApiKey = source["NotionApiKey"];
	        this.OpenAIApiKey = source["OpenAIApiKey"];
	        this.GroqApiKey = source["GroqApiKey"];
	        this.LocalWhisperModel = source["LocalWhisperModel"];
	        this.LocalWhisperGPU = source["LocalWhisperGPU"];
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
	export class GoogleAuthToken {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    // Go type: datatypes
	    user_info: any;
	    // Go type: datatypes
	    auth_token: any;
	
	    static createFrom(source: any = {}) {
	        return new GoogleAuthToken(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.user_info = this.convertValues(source["user_info"], null);
	        this.auth_token = this.convertValues(source["auth_token"], null);
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
	export class Page {
	    ID: string;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    title: string;
	    html_content: string;
	    blocks: number[];
	    is_folder: boolean;
	    expanded: boolean;
	    order: number;
	    ParentID?: string;
	    Children: Page[];
	
	    static createFrom(source: any = {}) {
	        return new Page(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.title = source["title"];
	        this.html_content = source["html_content"];
	        this.blocks = source["blocks"];
	        this.is_folder = source["is_folder"];
	        this.expanded = source["expanded"];
	        this.order = source["order"];
	        this.ParentID = source["ParentID"];
	        this.Children = this.convertValues(source["Children"], Page);
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

export namespace stt {
	
	export class Device {
	    Name: string;
	    IsDefault: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Device(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.IsDefault = source["IsDefault"];
	    }
	}

}

