// 小程序 COS 直传封装。
// 走「后端签发 STS 临时密钥 → 前端拿密钥直传 COS」的主流方案：
// 图片二进制不经过后端 / callContainer，因此没有包体大小限制（拍照/相册大图都能传）。
// #ifdef MP-WEIXIN
import COS from 'cos-wx-sdk-v5'
// #endif
import { api } from './index'
import { COS_BUCKET, COS_REGION } from '../config'

let cos = null

// 懒初始化 COS 实例：getAuthorization 每次上传前向后端要一张临时密钥，
// SDK 会自己缓存未过期的密钥，不会每次都打后端。
function ensureCos() {
  if (cos) return cos
  // #ifdef MP-WEIXIN
  cos = new COS({
    getAuthorization: (options, callback) => {
      api
        .cosCredential()
        .then((cred) => {
          callback({
            TmpSecretId: cred.tmpSecretId,
            TmpSecretKey: cred.tmpSecretKey,
            SecurityToken: cred.sessionToken,
            StartTime: cred.startTime,
            ExpiredTime: cred.expiredTime,
          })
        })
        .catch((e) => {
          // 传空密钥会让 SDK 抛错，进而走到 postObject 的 fail
          callback({ TmpSecretId: '', TmpSecretKey: '' })
          console.warn('🐶 获取 COS 临时密钥失败', e)
        })
    },
  })
  // #endif
  return cos
}

/**
 * 直传一张本地图片到 COS，resolve 出可公开访问的 https URL。
 * @param {string} filePath  chooseImage 拿到的本地临时路径
 * @param {string} ext       扩展名（含点），如 ".jpg"
 * @param {number} userId    仅用于生成不易碰撞的文件名
 */
export function uploadToCos(filePath, ext, userId) {
  return new Promise((resolve, reject) => {
    // #ifdef MP-WEIXIN
    const c = ensureCos()
    const rand = Math.floor(Math.random() * 1e6)
    const key = `munch/${userId || 0}_${Date.now()}_${rand}${ext || '.jpg'}`
    c.postObject(
      {
        Bucket: COS_BUCKET,
        Region: COS_REGION,
        Key: key,
        FilePath: filePath,
      },
      (err, data) => {
        if (err) {
          console.warn('🐶 COS 上传失败', err)
          reject(new Error('上传失败'))
          return
        }
        // data.Location 形如 ares1-xxx.cos.ap-beijing.myqcloud.com/munch/xxx.jpg（无协议）
        const url =
          data && data.Location
            ? 'https://' + data.Location.replace(/^https?:\/\//, '')
            : ''
        console.log('🐶 COS 上传成功', url)
        resolve(url)
      }
    )
    return
    // #endif
    // 非微信端不该走到这（H5 仍用 multipart uploadFile）
    reject(new Error('当前环境不支持 COS 直传'))
  })
}
