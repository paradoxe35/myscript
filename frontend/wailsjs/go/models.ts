export namespace local_whisper {
	
	export class DownloadProgress {
	    Name: string;
	    Size: number;
	    Total: number;
	
	    static createFrom(source: any = {}) {
	        return new DownloadProgress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Size = source["Size"];
	        this.Total = source["Total"];
	    }
	}
	export class LocalWhisperModel {
	    Name: string;
	    EnglishOnly: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LocalWhisperModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.EnglishOnly = source["EnglishOnly"];
	    }
	}

}

export namespace microphone {
	
	export class MicInputDevice {
	    Name: string;
	    IsDefault: number;
	    ID: any;
	
	    static createFrom(source: any = {}) {
	        return new MicInputDevice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.IsDefault = source["IsDefault"];
	        this.ID = source["ID"];
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
	    NotionApiKey?: string;
	    OpenAIApiKey?: string;
	    TranscriberSource: string;
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
	        this.NotionApiKey = source["NotionApiKey"];
	        this.OpenAIApiKey = source["OpenAIApiKey"];
	        this.TranscriberSource = source["TranscriberSource"];
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
	export class Page {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    title: string;
	    blocks: number[];
	
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
	        this.blocks = source["blocks"];
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

export namespace structs {
	
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

export namespace whisper {
	
	export class WhisperModel {
	    Name: string;
	    HasAlsoAnEnglishOnlyModel: boolean;
	    RAMRequired: number;
	    Enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WhisperModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.HasAlsoAnEnglishOnlyModel = source["HasAlsoAnEnglishOnlyModel"];
	        this.RAMRequired = source["RAMRequired"];
	        this.Enabled = source["Enabled"];
	    }
	}

}

