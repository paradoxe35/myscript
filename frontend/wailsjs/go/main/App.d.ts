// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {local_whisper} from '../models';
import {repository} from '../models';
import {structs} from '../models';
import {whisper} from '../models';
import {microphone} from '../models';
import {notion} from '../models';
import {notionapi} from '../models';

export function AreSomeLocalWhisperModelsDownloading():Promise<boolean>;

export function DeleteCache(arg1:string):Promise<void>;

export function DeleteLocalPage(arg1:number):Promise<void>;

export function DownloadLocalWhisperModels(arg1:Array<local_whisper.LocalWhisperModel>):Promise<void>;

export function ExistsLocalWhisperModel(arg1:local_whisper.LocalWhisperModel):Promise<boolean>;

export function GetAppVersion():Promise<string>;

export function GetBestLocalWhisperModel():Promise<string>;

export function GetCache(arg1:string):Promise<repository.CacheValue>;

export function GetConfig():Promise<repository.Config>;

export function GetLanguages():Promise<Array<structs.Language>>;

export function GetLocalPage(arg1:number):Promise<repository.Page>;

export function GetLocalPages():Promise<Array<repository.Page>>;

export function GetLocalWhisperDownloadProgress():Promise<local_whisper.DownloadProgress>;

export function GetLocalWhisperModels():Promise<Array<whisper.WhisperModel>>;

export function GetMicInputDevices():Promise<Array<microphone.MicInputDevice>>;

export function GetNotionPageBlocks(arg1:string):Promise<Array<notion.NotionBlock>>;

export function GetNotionPages():Promise<Array<notionapi.Object>>;

export function GetWhisperLanguages():Promise<Array<structs.Language>>;

export function GetWitAILanguages():Promise<Array<structs.Language>>;

export function IsLocalWhisperModelDownloading(arg1:local_whisper.LocalWhisperModel):Promise<boolean>;

export function IsRecording():Promise<boolean>;

export function LocalTranscribe(arg1:Array<number>,arg2:string):Promise<string>;

export function OpenAPITranscribe(arg1:Array<number>,arg2:string):Promise<string>;

export function SaveCache(arg1:string,arg2:any):Promise<repository.Cache>;

export function SaveConfig(arg1:repository.Config):Promise<repository.Config>;

export function SaveLocalPage(arg1:repository.Page):Promise<repository.Page>;

export function StartRecording(arg1:string,arg2:string):Promise<void>;

export function StopRecording():Promise<void>;

export function Transcribe(arg1:Array<number>,arg2:string):Promise<string>;

export function WitTranscribe(arg1:Array<number>,arg2:string):Promise<string>;
