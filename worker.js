"use strict";
(() => {
  var __create = Object.create;
  var __defProp = Object.defineProperty;
  var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
  var __getOwnPropNames = Object.getOwnPropertyNames;
  var __getProtoOf = Object.getPrototypeOf;
  var __hasOwnProp = Object.prototype.hasOwnProperty;
  var __require = /* @__PURE__ */ ((x) => typeof require !== "undefined" ? require : typeof Proxy !== "undefined" ? new Proxy(x, {
    get: (a, b) => (typeof require !== "undefined" ? require : a)[b]
  }) : x)(function(x) {
    if (typeof require !== "undefined")
      return require.apply(this, arguments);
    throw new Error('Dynamic require of "' + x + '" is not supported');
  });
  var __commonJS = (cb, mod) => function __require2() {
    return mod || (0, cb[__getOwnPropNames(cb)[0]])((mod = { exports: {} }).exports, mod), mod.exports;
  };
  var __copyProps = (to, from, except, desc) => {
    if (from && typeof from === "object" || typeof from === "function") {
      for (let key of __getOwnPropNames(from))
        if (!__hasOwnProp.call(to, key) && key !== except)
          __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
    }
    return to;
  };
  var __toESM = (mod, isNodeMode, target) => (target = mod != null ? __create(__getProtoOf(mod)) : {}, __copyProps(
    isNodeMode || !mod || !mod.__esModule ? __defProp(target, "default", { value: mod, enumerable: true }) : target,
    mod
  ));

  // node_modules/shopware-app-server-sdk/errors/ApiClient.js
  var require_ApiClient = __commonJS({
    "node_modules/shopware-app-server-sdk/errors/ApiClient.js"(exports) {
      "use strict";
      Object.defineProperty(exports, "__esModule", { value: true });
      exports.ApiClientRequestFailed = exports.ApiClientAuthenticationFailed = void 0;
      var ApiClientAuthenticationFailed = class extends Error {
        constructor(shopId, response) {
          super(`The api client authentication to shop with id: ${shopId}`);
          this.response = response;
        }
      };
      exports.ApiClientAuthenticationFailed = ApiClientAuthenticationFailed;
      var ApiClientRequestFailed = class extends Error {
        constructor(shopId, response) {
          super(`The api request failed with status code: ${response.statusCode} for shop with id: ${shopId}`);
          this.response = response;
        }
      };
      exports.ApiClientRequestFailed = ApiClientRequestFailed;
    }
  });

  // node_modules/shopware-app-server-sdk/component/http-client.js
  var require_http_client = __commonJS({
    "node_modules/shopware-app-server-sdk/component/http-client.js"(exports) {
      "use strict";
      Object.defineProperty(exports, "__esModule", { value: true });
      exports.HttpResponse = exports.HttpClient = void 0;
      var ApiClient_1 = require_ApiClient();
      var HttpClient3 = class {
        constructor(shop) {
          this.shop = shop;
          this.storage = {
            token: null,
            expiresIn: null
          };
        }
        async get(url, headers = {}) {
          return await this.request("GET", url, null, headers);
        }
        async post(url, json = {}, headers = {}) {
          headers["content-type"] = "application/json";
          headers["accept"] = "application/json";
          return await this.request("POST", url, JSON.stringify(json), headers);
        }
        async put(url, json = {}, headers = {}) {
          headers["content-type"] = "application/json";
          headers["accept"] = "application/json";
          return await this.request("PUT", url, JSON.stringify(json), headers);
        }
        async patch(url, json = {}, headers = {}) {
          headers["content-type"] = "application/json";
          headers["accept"] = "application/json";
          return await this.request("PATCH", url, JSON.stringify(json), headers);
        }
        async delete(url, json = {}, headers = {}) {
          headers["content-type"] = "application/json";
          headers["accept"] = "application/json";
          return await this.request("DELETE", url, JSON.stringify(json), headers);
        }
        async request(method, url, body = "", headers = {}) {
          const fHeaders = Object.assign({}, headers);
          fHeaders["Authorization"] = `Bearer ${await this.getToken()}`;
          const f = await globalThis.fetch(`${this.shop.shopUrl}/api${url}`, {
            body,
            headers: fHeaders,
            method
          });
          if (!f.ok && f.status === 401) {
            this.storage.expiresIn = null;
            return await this.request(method, url, body, headers);
          } else if (!f.ok) {
            throw new ApiClient_1.ApiClientRequestFailed(this.shop.id, new HttpResponse(f.status, await f.json(), f.headers));
          }
          if (f.status === 204) {
            return new HttpResponse(f.status, {}, f.headers);
          }
          return new HttpResponse(f.status, await f.json(), f.headers);
        }
        async getToken() {
          if (this.storage.expiresIn === null) {
            const auth = await globalThis.fetch(`${this.shop.shopUrl}/api/oauth/token`, {
              method: "POST",
              headers: {
                "content-type": "application/json"
              },
              body: JSON.stringify({
                grant_type: "client_credentials",
                client_id: this.shop.clientId,
                client_secret: this.shop.clientSecret
              })
            });
            if (!auth.ok) {
              throw new ApiClient_1.ApiClientAuthenticationFailed(this.shop.id, new HttpResponse(auth.status, await auth.json(), auth.headers));
            }
            const expireDate = new Date();
            const authBody = await auth.json();
            this.storage.token = authBody.access_token;
            expireDate.setSeconds(expireDate.getSeconds() + authBody.expires_in);
            this.storage.expiresIn = expireDate;
            return this.storage.token;
          }
          if (this.storage.expiresIn.getTime() < new Date().getTime()) {
            this.storage.expiresIn = null;
            return await this.getToken();
          }
          return this.storage.token;
        }
      };
      exports.HttpClient = HttpClient3;
      var HttpResponse = class {
        constructor(statusCode, body, headers) {
          this.statusCode = statusCode;
          this.body = body;
          this.headers = headers;
        }
      };
      exports.HttpResponse = HttpResponse;
    }
  });

  // node_modules/shopware-app-server-sdk/shop.js
  var require_shop = __commonJS({
    "node_modules/shopware-app-server-sdk/shop.js"(exports) {
      "use strict";
      Object.defineProperty(exports, "__esModule", { value: true });
      exports.Shop = void 0;
      var Shop3 = class {
        constructor(id, shopUrl, shopSecret, clientId = null, clientSecret = null, customFields = {}) {
          this.id = id;
          this.shopUrl = shopUrl;
          this.shopSecret = shopSecret;
          this.clientId = clientId;
          this.clientSecret = clientSecret;
          this.customFields = customFields;
        }
      };
      exports.Shop = Shop3;
    }
  });

  // node_modules/bcryptjs/dist/bcrypt.js
  var require_bcrypt = __commonJS({
    "node_modules/bcryptjs/dist/bcrypt.js"(exports, module) {
      (function(global, factory) {
        if (typeof define === "function" && define["amd"])
          define([], factory);
        else if (typeof __require === "function" && typeof module === "object" && module && module["exports"])
          module["exports"] = factory();
        else
          (global["dcodeIO"] = global["dcodeIO"] || {})["bcrypt"] = factory();
      })(exports, function() {
        "use strict";
        var bcrypt = {};
        var randomFallback = null;
        function random(len) {
          if (typeof module !== "undefined" && module && module["exports"])
            try {
              return __require("crypto")["randomBytes"](len);
            } catch (e2) {
            }
          try {
            var a;
            (self["crypto"] || self["msCrypto"])["getRandomValues"](a = new Uint32Array(len));
            return Array.prototype.slice.call(a);
          } catch (e2) {
          }
          if (!randomFallback)
            throw Error("Neither WebCryptoAPI nor a crypto module is available. Use bcrypt.setRandomFallback to set an alternative");
          return randomFallback(len);
        }
        var randomAvailable = false;
        try {
          random(1);
          randomAvailable = true;
        } catch (e2) {
        }
        randomFallback = null;
        bcrypt.setRandomFallback = function(random2) {
          randomFallback = random2;
        };
        bcrypt.genSaltSync = function(rounds, seed_length) {
          rounds = rounds || GENSALT_DEFAULT_LOG2_ROUNDS;
          if (typeof rounds !== "number")
            throw Error("Illegal arguments: " + typeof rounds + ", " + typeof seed_length);
          if (rounds < 4)
            rounds = 4;
          else if (rounds > 31)
            rounds = 31;
          var salt = [];
          salt.push("$2a$");
          if (rounds < 10)
            salt.push("0");
          salt.push(rounds.toString());
          salt.push("$");
          salt.push(base64_encode(random(BCRYPT_SALT_LEN), BCRYPT_SALT_LEN));
          return salt.join("");
        };
        bcrypt.genSalt = function(rounds, seed_length, callback) {
          if (typeof seed_length === "function")
            callback = seed_length, seed_length = void 0;
          if (typeof rounds === "function")
            callback = rounds, rounds = void 0;
          if (typeof rounds === "undefined")
            rounds = GENSALT_DEFAULT_LOG2_ROUNDS;
          else if (typeof rounds !== "number")
            throw Error("illegal arguments: " + typeof rounds);
          function _async(callback2) {
            nextTick(function() {
              try {
                callback2(null, bcrypt.genSaltSync(rounds));
              } catch (err) {
                callback2(err);
              }
            });
          }
          if (callback) {
            if (typeof callback !== "function")
              throw Error("Illegal callback: " + typeof callback);
            _async(callback);
          } else
            return new Promise(function(resolve, reject) {
              _async(function(err, res) {
                if (err) {
                  reject(err);
                  return;
                }
                resolve(res);
              });
            });
        };
        bcrypt.hashSync = function(s, salt) {
          if (typeof salt === "undefined")
            salt = GENSALT_DEFAULT_LOG2_ROUNDS;
          if (typeof salt === "number")
            salt = bcrypt.genSaltSync(salt);
          if (typeof s !== "string" || typeof salt !== "string")
            throw Error("Illegal arguments: " + typeof s + ", " + typeof salt);
          return _hash(s, salt);
        };
        bcrypt.hash = function(s, salt, callback, progressCallback) {
          function _async(callback2) {
            if (typeof s === "string" && typeof salt === "number")
              bcrypt.genSalt(salt, function(err, salt2) {
                _hash(s, salt2, callback2, progressCallback);
              });
            else if (typeof s === "string" && typeof salt === "string")
              _hash(s, salt, callback2, progressCallback);
            else
              nextTick(callback2.bind(this, Error("Illegal arguments: " + typeof s + ", " + typeof salt)));
          }
          if (callback) {
            if (typeof callback !== "function")
              throw Error("Illegal callback: " + typeof callback);
            _async(callback);
          } else
            return new Promise(function(resolve, reject) {
              _async(function(err, res) {
                if (err) {
                  reject(err);
                  return;
                }
                resolve(res);
              });
            });
        };
        function safeStringCompare(known, unknown) {
          var right = 0, wrong = 0;
          for (var i = 0, k = known.length; i < k; ++i) {
            if (known.charCodeAt(i) === unknown.charCodeAt(i))
              ++right;
            else
              ++wrong;
          }
          if (right < 0)
            return false;
          return wrong === 0;
        }
        bcrypt.compareSync = function(s, hash) {
          if (typeof s !== "string" || typeof hash !== "string")
            throw Error("Illegal arguments: " + typeof s + ", " + typeof hash);
          if (hash.length !== 60)
            return false;
          return safeStringCompare(bcrypt.hashSync(s, hash.substr(0, hash.length - 31)), hash);
        };
        bcrypt.compare = function(s, hash, callback, progressCallback) {
          function _async(callback2) {
            if (typeof s !== "string" || typeof hash !== "string") {
              nextTick(callback2.bind(this, Error("Illegal arguments: " + typeof s + ", " + typeof hash)));
              return;
            }
            if (hash.length !== 60) {
              nextTick(callback2.bind(this, null, false));
              return;
            }
            bcrypt.hash(s, hash.substr(0, 29), function(err, comp) {
              if (err)
                callback2(err);
              else
                callback2(null, safeStringCompare(comp, hash));
            }, progressCallback);
          }
          if (callback) {
            if (typeof callback !== "function")
              throw Error("Illegal callback: " + typeof callback);
            _async(callback);
          } else
            return new Promise(function(resolve, reject) {
              _async(function(err, res) {
                if (err) {
                  reject(err);
                  return;
                }
                resolve(res);
              });
            });
        };
        bcrypt.getRounds = function(hash) {
          if (typeof hash !== "string")
            throw Error("Illegal arguments: " + typeof hash);
          return parseInt(hash.split("$")[2], 10);
        };
        bcrypt.getSalt = function(hash) {
          if (typeof hash !== "string")
            throw Error("Illegal arguments: " + typeof hash);
          if (hash.length !== 60)
            throw Error("Illegal hash length: " + hash.length + " != 60");
          return hash.substring(0, 29);
        };
        var nextTick = typeof process !== "undefined" && process && typeof process.nextTick === "function" ? typeof setImmediate === "function" ? setImmediate : process.nextTick : setTimeout;
        function stringToBytes(str) {
          var out = [], i = 0;
          utfx.encodeUTF16toUTF8(function() {
            if (i >= str.length)
              return null;
            return str.charCodeAt(i++);
          }, function(b) {
            out.push(b);
          });
          return out;
        }
        var BASE64_CODE = "./ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789".split("");
        var BASE64_INDEX = [
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          0,
          1,
          54,
          55,
          56,
          57,
          58,
          59,
          60,
          61,
          62,
          63,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          2,
          3,
          4,
          5,
          6,
          7,
          8,
          9,
          10,
          11,
          12,
          13,
          14,
          15,
          16,
          17,
          18,
          19,
          20,
          21,
          22,
          23,
          24,
          25,
          26,
          27,
          -1,
          -1,
          -1,
          -1,
          -1,
          -1,
          28,
          29,
          30,
          31,
          32,
          33,
          34,
          35,
          36,
          37,
          38,
          39,
          40,
          41,
          42,
          43,
          44,
          45,
          46,
          47,
          48,
          49,
          50,
          51,
          52,
          53,
          -1,
          -1,
          -1,
          -1,
          -1
        ];
        var stringFromCharCode = String.fromCharCode;
        function base64_encode(b, len) {
          var off = 0, rs = [], c1, c2;
          if (len <= 0 || len > b.length)
            throw Error("Illegal len: " + len);
          while (off < len) {
            c1 = b[off++] & 255;
            rs.push(BASE64_CODE[c1 >> 2 & 63]);
            c1 = (c1 & 3) << 4;
            if (off >= len) {
              rs.push(BASE64_CODE[c1 & 63]);
              break;
            }
            c2 = b[off++] & 255;
            c1 |= c2 >> 4 & 15;
            rs.push(BASE64_CODE[c1 & 63]);
            c1 = (c2 & 15) << 2;
            if (off >= len) {
              rs.push(BASE64_CODE[c1 & 63]);
              break;
            }
            c2 = b[off++] & 255;
            c1 |= c2 >> 6 & 3;
            rs.push(BASE64_CODE[c1 & 63]);
            rs.push(BASE64_CODE[c2 & 63]);
          }
          return rs.join("");
        }
        function base64_decode(s, len) {
          var off = 0, slen = s.length, olen = 0, rs = [], c1, c2, c3, c4, o, code;
          if (len <= 0)
            throw Error("Illegal len: " + len);
          while (off < slen - 1 && olen < len) {
            code = s.charCodeAt(off++);
            c1 = code < BASE64_INDEX.length ? BASE64_INDEX[code] : -1;
            code = s.charCodeAt(off++);
            c2 = code < BASE64_INDEX.length ? BASE64_INDEX[code] : -1;
            if (c1 == -1 || c2 == -1)
              break;
            o = c1 << 2 >>> 0;
            o |= (c2 & 48) >> 4;
            rs.push(stringFromCharCode(o));
            if (++olen >= len || off >= slen)
              break;
            code = s.charCodeAt(off++);
            c3 = code < BASE64_INDEX.length ? BASE64_INDEX[code] : -1;
            if (c3 == -1)
              break;
            o = (c2 & 15) << 4 >>> 0;
            o |= (c3 & 60) >> 2;
            rs.push(stringFromCharCode(o));
            if (++olen >= len || off >= slen)
              break;
            code = s.charCodeAt(off++);
            c4 = code < BASE64_INDEX.length ? BASE64_INDEX[code] : -1;
            o = (c3 & 3) << 6 >>> 0;
            o |= c4;
            rs.push(stringFromCharCode(o));
            ++olen;
          }
          var res = [];
          for (off = 0; off < olen; off++)
            res.push(rs[off].charCodeAt(0));
          return res;
        }
        var utfx = function() {
          "use strict";
          var utfx2 = {};
          utfx2.MAX_CODEPOINT = 1114111;
          utfx2.encodeUTF8 = function(src, dst) {
            var cp = null;
            if (typeof src === "number")
              cp = src, src = function() {
                return null;
              };
            while (cp !== null || (cp = src()) !== null) {
              if (cp < 128)
                dst(cp & 127);
              else if (cp < 2048)
                dst(cp >> 6 & 31 | 192), dst(cp & 63 | 128);
              else if (cp < 65536)
                dst(cp >> 12 & 15 | 224), dst(cp >> 6 & 63 | 128), dst(cp & 63 | 128);
              else
                dst(cp >> 18 & 7 | 240), dst(cp >> 12 & 63 | 128), dst(cp >> 6 & 63 | 128), dst(cp & 63 | 128);
              cp = null;
            }
          };
          utfx2.decodeUTF8 = function(src, dst) {
            var a, b, c, d, fail = function(b2) {
              b2 = b2.slice(0, b2.indexOf(null));
              var err = Error(b2.toString());
              err.name = "TruncatedError";
              err["bytes"] = b2;
              throw err;
            };
            while ((a = src()) !== null) {
              if ((a & 128) === 0)
                dst(a);
              else if ((a & 224) === 192)
                (b = src()) === null && fail([a, b]), dst((a & 31) << 6 | b & 63);
              else if ((a & 240) === 224)
                ((b = src()) === null || (c = src()) === null) && fail([a, b, c]), dst((a & 15) << 12 | (b & 63) << 6 | c & 63);
              else if ((a & 248) === 240)
                ((b = src()) === null || (c = src()) === null || (d = src()) === null) && fail([a, b, c, d]), dst((a & 7) << 18 | (b & 63) << 12 | (c & 63) << 6 | d & 63);
              else
                throw RangeError("Illegal starting byte: " + a);
            }
          };
          utfx2.UTF16toUTF8 = function(src, dst) {
            var c1, c2 = null;
            while (true) {
              if ((c1 = c2 !== null ? c2 : src()) === null)
                break;
              if (c1 >= 55296 && c1 <= 57343) {
                if ((c2 = src()) !== null) {
                  if (c2 >= 56320 && c2 <= 57343) {
                    dst((c1 - 55296) * 1024 + c2 - 56320 + 65536);
                    c2 = null;
                    continue;
                  }
                }
              }
              dst(c1);
            }
            if (c2 !== null)
              dst(c2);
          };
          utfx2.UTF8toUTF16 = function(src, dst) {
            var cp = null;
            if (typeof src === "number")
              cp = src, src = function() {
                return null;
              };
            while (cp !== null || (cp = src()) !== null) {
              if (cp <= 65535)
                dst(cp);
              else
                cp -= 65536, dst((cp >> 10) + 55296), dst(cp % 1024 + 56320);
              cp = null;
            }
          };
          utfx2.encodeUTF16toUTF8 = function(src, dst) {
            utfx2.UTF16toUTF8(src, function(cp) {
              utfx2.encodeUTF8(cp, dst);
            });
          };
          utfx2.decodeUTF8toUTF16 = function(src, dst) {
            utfx2.decodeUTF8(src, function(cp) {
              utfx2.UTF8toUTF16(cp, dst);
            });
          };
          utfx2.calculateCodePoint = function(cp) {
            return cp < 128 ? 1 : cp < 2048 ? 2 : cp < 65536 ? 3 : 4;
          };
          utfx2.calculateUTF8 = function(src) {
            var cp, l = 0;
            while ((cp = src()) !== null)
              l += utfx2.calculateCodePoint(cp);
            return l;
          };
          utfx2.calculateUTF16asUTF8 = function(src) {
            var n = 0, l = 0;
            utfx2.UTF16toUTF8(src, function(cp) {
              ++n;
              l += utfx2.calculateCodePoint(cp);
            });
            return [n, l];
          };
          return utfx2;
        }();
        Date.now = Date.now || function() {
          return +new Date();
        };
        var BCRYPT_SALT_LEN = 16;
        var GENSALT_DEFAULT_LOG2_ROUNDS = 10;
        var BLOWFISH_NUM_ROUNDS = 16;
        var MAX_EXECUTION_TIME = 100;
        var P_ORIG = [
          608135816,
          2242054355,
          320440878,
          57701188,
          2752067618,
          698298832,
          137296536,
          3964562569,
          1160258022,
          953160567,
          3193202383,
          887688300,
          3232508343,
          3380367581,
          1065670069,
          3041331479,
          2450970073,
          2306472731
        ];
        var S_ORIG = [
          3509652390,
          2564797868,
          805139163,
          3491422135,
          3101798381,
          1780907670,
          3128725573,
          4046225305,
          614570311,
          3012652279,
          134345442,
          2240740374,
          1667834072,
          1901547113,
          2757295779,
          4103290238,
          227898511,
          1921955416,
          1904987480,
          2182433518,
          2069144605,
          3260701109,
          2620446009,
          720527379,
          3318853667,
          677414384,
          3393288472,
          3101374703,
          2390351024,
          1614419982,
          1822297739,
          2954791486,
          3608508353,
          3174124327,
          2024746970,
          1432378464,
          3864339955,
          2857741204,
          1464375394,
          1676153920,
          1439316330,
          715854006,
          3033291828,
          289532110,
          2706671279,
          2087905683,
          3018724369,
          1668267050,
          732546397,
          1947742710,
          3462151702,
          2609353502,
          2950085171,
          1814351708,
          2050118529,
          680887927,
          999245976,
          1800124847,
          3300911131,
          1713906067,
          1641548236,
          4213287313,
          1216130144,
          1575780402,
          4018429277,
          3917837745,
          3693486850,
          3949271944,
          596196993,
          3549867205,
          258830323,
          2213823033,
          772490370,
          2760122372,
          1774776394,
          2652871518,
          566650946,
          4142492826,
          1728879713,
          2882767088,
          1783734482,
          3629395816,
          2517608232,
          2874225571,
          1861159788,
          326777828,
          3124490320,
          2130389656,
          2716951837,
          967770486,
          1724537150,
          2185432712,
          2364442137,
          1164943284,
          2105845187,
          998989502,
          3765401048,
          2244026483,
          1075463327,
          1455516326,
          1322494562,
          910128902,
          469688178,
          1117454909,
          936433444,
          3490320968,
          3675253459,
          1240580251,
          122909385,
          2157517691,
          634681816,
          4142456567,
          3825094682,
          3061402683,
          2540495037,
          79693498,
          3249098678,
          1084186820,
          1583128258,
          426386531,
          1761308591,
          1047286709,
          322548459,
          995290223,
          1845252383,
          2603652396,
          3431023940,
          2942221577,
          3202600964,
          3727903485,
          1712269319,
          422464435,
          3234572375,
          1170764815,
          3523960633,
          3117677531,
          1434042557,
          442511882,
          3600875718,
          1076654713,
          1738483198,
          4213154764,
          2393238008,
          3677496056,
          1014306527,
          4251020053,
          793779912,
          2902807211,
          842905082,
          4246964064,
          1395751752,
          1040244610,
          2656851899,
          3396308128,
          445077038,
          3742853595,
          3577915638,
          679411651,
          2892444358,
          2354009459,
          1767581616,
          3150600392,
          3791627101,
          3102740896,
          284835224,
          4246832056,
          1258075500,
          768725851,
          2589189241,
          3069724005,
          3532540348,
          1274779536,
          3789419226,
          2764799539,
          1660621633,
          3471099624,
          4011903706,
          913787905,
          3497959166,
          737222580,
          2514213453,
          2928710040,
          3937242737,
          1804850592,
          3499020752,
          2949064160,
          2386320175,
          2390070455,
          2415321851,
          4061277028,
          2290661394,
          2416832540,
          1336762016,
          1754252060,
          3520065937,
          3014181293,
          791618072,
          3188594551,
          3933548030,
          2332172193,
          3852520463,
          3043980520,
          413987798,
          3465142937,
          3030929376,
          4245938359,
          2093235073,
          3534596313,
          375366246,
          2157278981,
          2479649556,
          555357303,
          3870105701,
          2008414854,
          3344188149,
          4221384143,
          3956125452,
          2067696032,
          3594591187,
          2921233993,
          2428461,
          544322398,
          577241275,
          1471733935,
          610547355,
          4027169054,
          1432588573,
          1507829418,
          2025931657,
          3646575487,
          545086370,
          48609733,
          2200306550,
          1653985193,
          298326376,
          1316178497,
          3007786442,
          2064951626,
          458293330,
          2589141269,
          3591329599,
          3164325604,
          727753846,
          2179363840,
          146436021,
          1461446943,
          4069977195,
          705550613,
          3059967265,
          3887724982,
          4281599278,
          3313849956,
          1404054877,
          2845806497,
          146425753,
          1854211946,
          1266315497,
          3048417604,
          3681880366,
          3289982499,
          290971e4,
          1235738493,
          2632868024,
          2414719590,
          3970600049,
          1771706367,
          1449415276,
          3266420449,
          422970021,
          1963543593,
          2690192192,
          3826793022,
          1062508698,
          1531092325,
          1804592342,
          2583117782,
          2714934279,
          4024971509,
          1294809318,
          4028980673,
          1289560198,
          2221992742,
          1669523910,
          35572830,
          157838143,
          1052438473,
          1016535060,
          1802137761,
          1753167236,
          1386275462,
          3080475397,
          2857371447,
          1040679964,
          2145300060,
          2390574316,
          1461121720,
          2956646967,
          4031777805,
          4028374788,
          33600511,
          2920084762,
          1018524850,
          629373528,
          3691585981,
          3515945977,
          2091462646,
          2486323059,
          586499841,
          988145025,
          935516892,
          3367335476,
          2599673255,
          2839830854,
          265290510,
          3972581182,
          2759138881,
          3795373465,
          1005194799,
          847297441,
          406762289,
          1314163512,
          1332590856,
          1866599683,
          4127851711,
          750260880,
          613907577,
          1450815602,
          3165620655,
          3734664991,
          3650291728,
          3012275730,
          3704569646,
          1427272223,
          778793252,
          1343938022,
          2676280711,
          2052605720,
          1946737175,
          3164576444,
          3914038668,
          3967478842,
          3682934266,
          1661551462,
          3294938066,
          4011595847,
          840292616,
          3712170807,
          616741398,
          312560963,
          711312465,
          1351876610,
          322626781,
          1910503582,
          271666773,
          2175563734,
          1594956187,
          70604529,
          3617834859,
          1007753275,
          1495573769,
          4069517037,
          2549218298,
          2663038764,
          504708206,
          2263041392,
          3941167025,
          2249088522,
          1514023603,
          1998579484,
          1312622330,
          694541497,
          2582060303,
          2151582166,
          1382467621,
          776784248,
          2618340202,
          3323268794,
          2497899128,
          2784771155,
          503983604,
          4076293799,
          907881277,
          423175695,
          432175456,
          1378068232,
          4145222326,
          3954048622,
          3938656102,
          3820766613,
          2793130115,
          2977904593,
          26017576,
          3274890735,
          3194772133,
          1700274565,
          1756076034,
          4006520079,
          3677328699,
          720338349,
          1533947780,
          354530856,
          688349552,
          3973924725,
          1637815568,
          332179504,
          3949051286,
          53804574,
          2852348879,
          3044236432,
          1282449977,
          3583942155,
          3416972820,
          4006381244,
          1617046695,
          2628476075,
          3002303598,
          1686838959,
          431878346,
          2686675385,
          1700445008,
          1080580658,
          1009431731,
          832498133,
          3223435511,
          2605976345,
          2271191193,
          2516031870,
          1648197032,
          4164389018,
          2548247927,
          300782431,
          375919233,
          238389289,
          3353747414,
          2531188641,
          2019080857,
          1475708069,
          455242339,
          2609103871,
          448939670,
          3451063019,
          1395535956,
          2413381860,
          1841049896,
          1491858159,
          885456874,
          4264095073,
          4001119347,
          1565136089,
          3898914787,
          1108368660,
          540939232,
          1173283510,
          2745871338,
          3681308437,
          4207628240,
          3343053890,
          4016749493,
          1699691293,
          1103962373,
          3625875870,
          2256883143,
          3830138730,
          1031889488,
          3479347698,
          1535977030,
          4236805024,
          3251091107,
          2132092099,
          1774941330,
          1199868427,
          1452454533,
          157007616,
          2904115357,
          342012276,
          595725824,
          1480756522,
          206960106,
          497939518,
          591360097,
          863170706,
          2375253569,
          3596610801,
          1814182875,
          2094937945,
          3421402208,
          1082520231,
          3463918190,
          2785509508,
          435703966,
          3908032597,
          1641649973,
          2842273706,
          3305899714,
          1510255612,
          2148256476,
          2655287854,
          3276092548,
          4258621189,
          236887753,
          3681803219,
          274041037,
          1734335097,
          3815195456,
          3317970021,
          1899903192,
          1026095262,
          4050517792,
          356393447,
          2410691914,
          3873677099,
          3682840055,
          3913112168,
          2491498743,
          4132185628,
          2489919796,
          1091903735,
          1979897079,
          3170134830,
          3567386728,
          3557303409,
          857797738,
          1136121015,
          1342202287,
          507115054,
          2535736646,
          337727348,
          3213592640,
          1301675037,
          2528481711,
          1895095763,
          1721773893,
          3216771564,
          62756741,
          2142006736,
          835421444,
          2531993523,
          1442658625,
          3659876326,
          2882144922,
          676362277,
          1392781812,
          170690266,
          3921047035,
          1759253602,
          3611846912,
          1745797284,
          664899054,
          1329594018,
          3901205900,
          3045908486,
          2062866102,
          2865634940,
          3543621612,
          3464012697,
          1080764994,
          553557557,
          3656615353,
          3996768171,
          991055499,
          499776247,
          1265440854,
          648242737,
          3940784050,
          980351604,
          3713745714,
          1749149687,
          3396870395,
          4211799374,
          3640570775,
          1161844396,
          3125318951,
          1431517754,
          545492359,
          4268468663,
          3499529547,
          1437099964,
          2702547544,
          3433638243,
          2581715763,
          2787789398,
          1060185593,
          1593081372,
          2418618748,
          4260947970,
          69676912,
          2159744348,
          86519011,
          2512459080,
          3838209314,
          1220612927,
          3339683548,
          133810670,
          1090789135,
          1078426020,
          1569222167,
          845107691,
          3583754449,
          4072456591,
          1091646820,
          628848692,
          1613405280,
          3757631651,
          526609435,
          236106946,
          48312990,
          2942717905,
          3402727701,
          1797494240,
          859738849,
          992217954,
          4005476642,
          2243076622,
          3870952857,
          3732016268,
          765654824,
          3490871365,
          2511836413,
          1685915746,
          3888969200,
          1414112111,
          2273134842,
          3281911079,
          4080962846,
          172450625,
          2569994100,
          980381355,
          4109958455,
          2819808352,
          2716589560,
          2568741196,
          3681446669,
          3329971472,
          1835478071,
          660984891,
          3704678404,
          4045999559,
          3422617507,
          3040415634,
          1762651403,
          1719377915,
          3470491036,
          2693910283,
          3642056355,
          3138596744,
          1364962596,
          2073328063,
          1983633131,
          926494387,
          3423689081,
          2150032023,
          4096667949,
          1749200295,
          3328846651,
          309677260,
          2016342300,
          1779581495,
          3079819751,
          111262694,
          1274766160,
          443224088,
          298511866,
          1025883608,
          3806446537,
          1145181785,
          168956806,
          3641502830,
          3584813610,
          1689216846,
          3666258015,
          3200248200,
          1692713982,
          2646376535,
          4042768518,
          1618508792,
          1610833997,
          3523052358,
          4130873264,
          2001055236,
          3610705100,
          2202168115,
          4028541809,
          2961195399,
          1006657119,
          2006996926,
          3186142756,
          1430667929,
          3210227297,
          1314452623,
          4074634658,
          4101304120,
          2273951170,
          1399257539,
          3367210612,
          3027628629,
          1190975929,
          2062231137,
          2333990788,
          2221543033,
          2438960610,
          1181637006,
          548689776,
          2362791313,
          3372408396,
          3104550113,
          3145860560,
          296247880,
          1970579870,
          3078560182,
          3769228297,
          1714227617,
          3291629107,
          3898220290,
          166772364,
          1251581989,
          493813264,
          448347421,
          195405023,
          2709975567,
          677966185,
          3703036547,
          1463355134,
          2715995803,
          1338867538,
          1343315457,
          2802222074,
          2684532164,
          233230375,
          2599980071,
          2000651841,
          3277868038,
          1638401717,
          4028070440,
          3237316320,
          6314154,
          819756386,
          300326615,
          590932579,
          1405279636,
          3267499572,
          3150704214,
          2428286686,
          3959192993,
          3461946742,
          1862657033,
          1266418056,
          963775037,
          2089974820,
          2263052895,
          1917689273,
          448879540,
          3550394620,
          3981727096,
          150775221,
          3627908307,
          1303187396,
          508620638,
          2975983352,
          2726630617,
          1817252668,
          1876281319,
          1457606340,
          908771278,
          3720792119,
          3617206836,
          2455994898,
          1729034894,
          1080033504,
          976866871,
          3556439503,
          2881648439,
          1522871579,
          1555064734,
          1336096578,
          3548522304,
          2579274686,
          3574697629,
          3205460757,
          3593280638,
          3338716283,
          3079412587,
          564236357,
          2993598910,
          1781952180,
          1464380207,
          3163844217,
          3332601554,
          1699332808,
          1393555694,
          1183702653,
          3581086237,
          1288719814,
          691649499,
          2847557200,
          2895455976,
          3193889540,
          2717570544,
          1781354906,
          1676643554,
          2592534050,
          3230253752,
          1126444790,
          2770207658,
          2633158820,
          2210423226,
          2615765581,
          2414155088,
          3127139286,
          673620729,
          2805611233,
          1269405062,
          4015350505,
          3341807571,
          4149409754,
          1057255273,
          2012875353,
          2162469141,
          2276492801,
          2601117357,
          993977747,
          3918593370,
          2654263191,
          753973209,
          36408145,
          2530585658,
          25011837,
          3520020182,
          2088578344,
          530523599,
          2918365339,
          1524020338,
          1518925132,
          3760827505,
          3759777254,
          1202760957,
          3985898139,
          3906192525,
          674977740,
          4174734889,
          2031300136,
          2019492241,
          3983892565,
          4153806404,
          3822280332,
          352677332,
          2297720250,
          60907813,
          90501309,
          3286998549,
          1016092578,
          2535922412,
          2839152426,
          457141659,
          509813237,
          4120667899,
          652014361,
          1966332200,
          2975202805,
          55981186,
          2327461051,
          676427537,
          3255491064,
          2882294119,
          3433927263,
          1307055953,
          942726286,
          933058658,
          2468411793,
          3933900994,
          4215176142,
          1361170020,
          2001714738,
          2830558078,
          3274259782,
          1222529897,
          1679025792,
          2729314320,
          3714953764,
          1770335741,
          151462246,
          3013232138,
          1682292957,
          1483529935,
          471910574,
          1539241949,
          458788160,
          3436315007,
          1807016891,
          3718408830,
          978976581,
          1043663428,
          3165965781,
          1927990952,
          4200891579,
          2372276910,
          3208408903,
          3533431907,
          1412390302,
          2931980059,
          4132332400,
          1947078029,
          3881505623,
          4168226417,
          2941484381,
          1077988104,
          1320477388,
          886195818,
          18198404,
          3786409e3,
          2509781533,
          112762804,
          3463356488,
          1866414978,
          891333506,
          18488651,
          661792760,
          1628790961,
          3885187036,
          3141171499,
          876946877,
          2693282273,
          1372485963,
          791857591,
          2686433993,
          3759982718,
          3167212022,
          3472953795,
          2716379847,
          445679433,
          3561995674,
          3504004811,
          3574258232,
          54117162,
          3331405415,
          2381918588,
          3769707343,
          4154350007,
          1140177722,
          4074052095,
          668550556,
          3214352940,
          367459370,
          261225585,
          2610173221,
          4209349473,
          3468074219,
          3265815641,
          314222801,
          3066103646,
          3808782860,
          282218597,
          3406013506,
          3773591054,
          379116347,
          1285071038,
          846784868,
          2669647154,
          3771962079,
          3550491691,
          2305946142,
          453669953,
          1268987020,
          3317592352,
          3279303384,
          3744833421,
          2610507566,
          3859509063,
          266596637,
          3847019092,
          517658769,
          3462560207,
          3443424879,
          370717030,
          4247526661,
          2224018117,
          4143653529,
          4112773975,
          2788324899,
          2477274417,
          1456262402,
          2901442914,
          1517677493,
          1846949527,
          2295493580,
          3734397586,
          2176403920,
          1280348187,
          1908823572,
          3871786941,
          846861322,
          1172426758,
          3287448474,
          3383383037,
          1655181056,
          3139813346,
          901632758,
          1897031941,
          2986607138,
          3066810236,
          3447102507,
          1393639104,
          373351379,
          950779232,
          625454576,
          3124240540,
          4148612726,
          2007998917,
          544563296,
          2244738638,
          2330496472,
          2058025392,
          1291430526,
          424198748,
          50039436,
          29584100,
          3605783033,
          2429876329,
          2791104160,
          1057563949,
          3255363231,
          3075367218,
          3463963227,
          1469046755,
          985887462
        ];
        var C_ORIG = [
          1332899944,
          1700884034,
          1701343084,
          1684370003,
          1668446532,
          1869963892
        ];
        function _encipher(lr, off, P, S) {
          var n, l = lr[off], r = lr[off + 1];
          l ^= P[0];
          n = S[l >>> 24];
          n += S[256 | l >> 16 & 255];
          n ^= S[512 | l >> 8 & 255];
          n += S[768 | l & 255];
          r ^= n ^ P[1];
          n = S[r >>> 24];
          n += S[256 | r >> 16 & 255];
          n ^= S[512 | r >> 8 & 255];
          n += S[768 | r & 255];
          l ^= n ^ P[2];
          n = S[l >>> 24];
          n += S[256 | l >> 16 & 255];
          n ^= S[512 | l >> 8 & 255];
          n += S[768 | l & 255];
          r ^= n ^ P[3];
          n = S[r >>> 24];
          n += S[256 | r >> 16 & 255];
          n ^= S[512 | r >> 8 & 255];
          n += S[768 | r & 255];
          l ^= n ^ P[4];
          n = S[l >>> 24];
          n += S[256 | l >> 16 & 255];
          n ^= S[512 | l >> 8 & 255];
          n += S[768 | l & 255];
          r ^= n ^ P[5];
          n = S[r >>> 24];
          n += S[256 | r >> 16 & 255];
          n ^= S[512 | r >> 8 & 255];
          n += S[768 | r & 255];
          l ^= n ^ P[6];
          n = S[l >>> 24];
          n += S[256 | l >> 16 & 255];
          n ^= S[512 | l >> 8 & 255];
          n += S[768 | l & 255];
          r ^= n ^ P[7];
          n = S[r >>> 24];
          n += S[256 | r >> 16 & 255];
          n ^= S[512 | r >> 8 & 255];
          n += S[768 | r & 255];
          l ^= n ^ P[8];
          n = S[l >>> 24];
          n += S[256 | l >> 16 & 255];
          n ^= S[512 | l >> 8 & 255];
          n += S[768 | l & 255];
          r ^= n ^ P[9];
          n = S[r >>> 24];
          n += S[256 | r >> 16 & 255];
          n ^= S[512 | r >> 8 & 255];
          n += S[768 | r & 255];
          l ^= n ^ P[10];
          n = S[l >>> 24];
          n += S[256 | l >> 16 & 255];
          n ^= S[512 | l >> 8 & 255];
          n += S[768 | l & 255];
          r ^= n ^ P[11];
          n = S[r >>> 24];
          n += S[256 | r >> 16 & 255];
          n ^= S[512 | r >> 8 & 255];
          n += S[768 | r & 255];
          l ^= n ^ P[12];
          n = S[l >>> 24];
          n += S[256 | l >> 16 & 255];
          n ^= S[512 | l >> 8 & 255];
          n += S[768 | l & 255];
          r ^= n ^ P[13];
          n = S[r >>> 24];
          n += S[256 | r >> 16 & 255];
          n ^= S[512 | r >> 8 & 255];
          n += S[768 | r & 255];
          l ^= n ^ P[14];
          n = S[l >>> 24];
          n += S[256 | l >> 16 & 255];
          n ^= S[512 | l >> 8 & 255];
          n += S[768 | l & 255];
          r ^= n ^ P[15];
          n = S[r >>> 24];
          n += S[256 | r >> 16 & 255];
          n ^= S[512 | r >> 8 & 255];
          n += S[768 | r & 255];
          l ^= n ^ P[16];
          lr[off] = r ^ P[BLOWFISH_NUM_ROUNDS + 1];
          lr[off + 1] = l;
          return lr;
        }
        function _streamtoword(data, offp) {
          for (var i = 0, word = 0; i < 4; ++i)
            word = word << 8 | data[offp] & 255, offp = (offp + 1) % data.length;
          return { key: word, offp };
        }
        function _key(key, P, S) {
          var offset = 0, lr = [0, 0], plen = P.length, slen = S.length, sw;
          for (var i = 0; i < plen; i++)
            sw = _streamtoword(key, offset), offset = sw.offp, P[i] = P[i] ^ sw.key;
          for (i = 0; i < plen; i += 2)
            lr = _encipher(lr, 0, P, S), P[i] = lr[0], P[i + 1] = lr[1];
          for (i = 0; i < slen; i += 2)
            lr = _encipher(lr, 0, P, S), S[i] = lr[0], S[i + 1] = lr[1];
        }
        function _ekskey(data, key, P, S) {
          var offp = 0, lr = [0, 0], plen = P.length, slen = S.length, sw;
          for (var i = 0; i < plen; i++)
            sw = _streamtoword(key, offp), offp = sw.offp, P[i] = P[i] ^ sw.key;
          offp = 0;
          for (i = 0; i < plen; i += 2)
            sw = _streamtoword(data, offp), offp = sw.offp, lr[0] ^= sw.key, sw = _streamtoword(data, offp), offp = sw.offp, lr[1] ^= sw.key, lr = _encipher(lr, 0, P, S), P[i] = lr[0], P[i + 1] = lr[1];
          for (i = 0; i < slen; i += 2)
            sw = _streamtoword(data, offp), offp = sw.offp, lr[0] ^= sw.key, sw = _streamtoword(data, offp), offp = sw.offp, lr[1] ^= sw.key, lr = _encipher(lr, 0, P, S), S[i] = lr[0], S[i + 1] = lr[1];
        }
        function _crypt(b, salt, rounds, callback, progressCallback) {
          var cdata = C_ORIG.slice(), clen = cdata.length, err;
          if (rounds < 4 || rounds > 31) {
            err = Error("Illegal number of rounds (4-31): " + rounds);
            if (callback) {
              nextTick(callback.bind(this, err));
              return;
            } else
              throw err;
          }
          if (salt.length !== BCRYPT_SALT_LEN) {
            err = Error("Illegal salt length: " + salt.length + " != " + BCRYPT_SALT_LEN);
            if (callback) {
              nextTick(callback.bind(this, err));
              return;
            } else
              throw err;
          }
          rounds = 1 << rounds >>> 0;
          var P, S, i = 0, j;
          if (Int32Array) {
            P = new Int32Array(P_ORIG);
            S = new Int32Array(S_ORIG);
          } else {
            P = P_ORIG.slice();
            S = S_ORIG.slice();
          }
          _ekskey(salt, b, P, S);
          function next() {
            if (progressCallback)
              progressCallback(i / rounds);
            if (i < rounds) {
              var start = Date.now();
              for (; i < rounds; ) {
                i = i + 1;
                _key(b, P, S);
                _key(salt, P, S);
                if (Date.now() - start > MAX_EXECUTION_TIME)
                  break;
              }
            } else {
              for (i = 0; i < 64; i++)
                for (j = 0; j < clen >> 1; j++)
                  _encipher(cdata, j << 1, P, S);
              var ret = [];
              for (i = 0; i < clen; i++)
                ret.push((cdata[i] >> 24 & 255) >>> 0), ret.push((cdata[i] >> 16 & 255) >>> 0), ret.push((cdata[i] >> 8 & 255) >>> 0), ret.push((cdata[i] & 255) >>> 0);
              if (callback) {
                callback(null, ret);
                return;
              } else
                return ret;
            }
            if (callback)
              nextTick(next);
          }
          if (typeof callback !== "undefined") {
            next();
          } else {
            var res;
            while (true)
              if (typeof (res = next()) !== "undefined")
                return res || [];
          }
        }
        function _hash(s, salt, callback, progressCallback) {
          var err;
          if (typeof s !== "string" || typeof salt !== "string") {
            err = Error("Invalid string / salt: Not a string");
            if (callback) {
              nextTick(callback.bind(this, err));
              return;
            } else
              throw err;
          }
          var minor, offset;
          if (salt.charAt(0) !== "$" || salt.charAt(1) !== "2") {
            err = Error("Invalid salt version: " + salt.substring(0, 2));
            if (callback) {
              nextTick(callback.bind(this, err));
              return;
            } else
              throw err;
          }
          if (salt.charAt(2) === "$")
            minor = String.fromCharCode(0), offset = 3;
          else {
            minor = salt.charAt(2);
            if (minor !== "a" && minor !== "b" && minor !== "y" || salt.charAt(3) !== "$") {
              err = Error("Invalid salt revision: " + salt.substring(2, 4));
              if (callback) {
                nextTick(callback.bind(this, err));
                return;
              } else
                throw err;
            }
            offset = 4;
          }
          if (salt.charAt(offset + 2) > "$") {
            err = Error("Missing salt rounds");
            if (callback) {
              nextTick(callback.bind(this, err));
              return;
            } else
              throw err;
          }
          var r1 = parseInt(salt.substring(offset, offset + 1), 10) * 10, r2 = parseInt(salt.substring(offset + 1, offset + 2), 10), rounds = r1 + r2, real_salt = salt.substring(offset + 3, offset + 25);
          s += minor >= "a" ? "\0" : "";
          var passwordb = stringToBytes(s), saltb = base64_decode(real_salt, BCRYPT_SALT_LEN);
          function finish(bytes2) {
            var res = [];
            res.push("$2");
            if (minor >= "a")
              res.push(minor);
            res.push("$");
            if (rounds < 10)
              res.push("0");
            res.push(rounds.toString());
            res.push("$");
            res.push(base64_encode(saltb, saltb.length));
            res.push(base64_encode(bytes2, C_ORIG.length * 4 - 1));
            return res.join("");
          }
          if (typeof callback == "undefined")
            return finish(_crypt(passwordb, saltb, rounds));
          else {
            _crypt(passwordb, saltb, rounds, function(err2, bytes2) {
              if (err2)
                callback(err2, null);
              else
                callback(null, finish(bytes2));
            }, progressCallback);
          }
        }
        bcrypt.encodeBase64 = base64_encode;
        bcrypt.decodeBase64 = base64_decode;
        return bcrypt;
      });
    }
  });

  // src/cron/schedule.ts
  var import_http_client = __toESM(require_http_client());
  var import_shop = __toESM(require_shop());

  // node_modules/@planetscale/database/dist/sanitization.js
  function format(query, values) {
    return Array.isArray(values) ? replacePosition(query, values) : replaceNamed(query, values);
  }
  function replacePosition(query, values) {
    let index = 0;
    return query.replace(/\?/g, (match) => {
      return index < values.length ? sanitize(values[index++]) : match;
    });
  }
  function replaceNamed(query, values) {
    return query.replace(/:(\w+)/g, (match, name) => {
      return hasOwn(values, name) ? sanitize(values[name]) : match;
    });
  }
  function hasOwn(obj, name) {
    return Object.prototype.hasOwnProperty.call(obj, name);
  }
  function sanitize(value) {
    if (value == null) {
      return "null";
    }
    if (typeof value === "number") {
      return String(value);
    }
    if (typeof value === "boolean") {
      return value ? "true" : "false";
    }
    if (typeof value === "string") {
      return quote(value);
    }
    if (Array.isArray(value)) {
      return value.map(sanitize).join(", ");
    }
    if (value instanceof Date) {
      return quote(value.toISOString());
    }
    return quote(value.toString());
  }
  function quote(text) {
    return `'${escape(text)}'`;
  }
  var re = /[\0\b\n\r\t\x1a\\"']/g;
  function escape(text) {
    return text.replace(re, replacement);
  }
  function replacement(text) {
    switch (text) {
      case '"':
        return '\\"';
      case "'":
        return "\\'";
      case "\n":
        return "\\n";
      case "\r":
        return "\\r";
      case "	":
        return "\\t";
      case "\\":
        return "\\\\";
      case "\0":
        return "\\0";
      case "\b":
        return "\\b";
      case "":
        return "\\Z";
      default:
        return "";
    }
  }

  // node_modules/@planetscale/database/dist/text.js
  var decoder = new TextDecoder("utf-8");
  function decode(text) {
    return text ? decoder.decode(Uint8Array.from(bytes(text))) : "";
  }
  function bytes(text) {
    return text.split("").map((c) => c.charCodeAt(0));
  }

  // node_modules/@planetscale/database/dist/version.js
  var Version = "1.3.0";

  // node_modules/@planetscale/database/dist/index.js
  var DatabaseError = class extends Error {
    constructor(message, status, body) {
      super(message);
      this.status = status;
      this.name = "DatabaseError";
      this.body = body;
    }
  };
  var Tx = class {
    constructor(conn) {
      this.conn = conn;
    }
    async execute(query, args) {
      return this.conn.execute(query, args);
    }
  };
  var Connection = class {
    constructor(config2) {
      var _a;
      this.session = null;
      this.config = { ...config2 };
      if (typeof fetch !== "undefined") {
        (_a = this.config).fetch || (_a.fetch = fetch);
      }
      if (config2.url) {
        const url = new URL(config2.url);
        this.config.username = url.username;
        this.config.password = url.password;
        this.config.host = url.hostname;
      }
    }
    async transaction(fn) {
      const conn = new Connection(this.config);
      const tx = new Tx(conn);
      try {
        await tx.execute("BEGIN");
        const res = await fn(tx);
        await tx.execute("COMMIT");
        return res;
      } catch (err) {
        await tx.execute("ROLLBACK");
        throw err;
      }
    }
    async refresh() {
      await this.createSession();
    }
    async execute(query, args) {
      const url = new URL("/psdb.v1alpha1.Database/Execute", `https://${this.config.host}`);
      const formatter = this.config.format || format;
      const sql = args ? formatter(query, args) : query;
      const start = Date.now();
      const saved = await postJSON(this.config, url, { query: sql, session: this.session });
      const time = Date.now() - start;
      const { result, session, error } = saved;
      if (error) {
        throw new DatabaseError(error.message, 400, error);
      }
      const rowsAffected = result?.rowsAffected ? parseInt(result.rowsAffected, 10) : null;
      const insertId = result?.insertId ?? null;
      this.session = session;
      const rows = result ? parse(result, this.config.cast || cast) : [];
      const headers = result ? result.fields?.map((f) => f.name) ?? [] : [];
      const typeByName = (acc, { name, type }) => ({ ...acc, [name]: type });
      const types = result ? result.fields?.reduce(typeByName, {}) ?? {} : {};
      return {
        headers,
        types,
        rows,
        rowsAffected,
        insertId,
        size: rows.length,
        statement: sql,
        time
      };
    }
    async createSession() {
      const url = new URL("/psdb.v1alpha1.Database/CreateSession", `https://${this.config.host}`);
      const { session } = await postJSON(this.config, url);
      this.session = session;
      return session;
    }
  };
  async function postJSON(config2, url, body = {}) {
    const auth = btoa(`${config2.username}:${config2.password}`);
    const { fetch: fetch2 } = config2;
    const response = await fetch2(url.toString(), {
      method: "POST",
      body: JSON.stringify(body),
      headers: {
        "Content-Type": "application/json",
        "User-Agent": `database-js/${Version}`,
        Authorization: `Basic ${auth}`
      }
    });
    if (response.ok) {
      return await response.json();
    } else {
      let error = null;
      try {
        const e2 = (await response.json()).error;
        error = new DatabaseError(e2.message, response.status, e2);
      } catch {
        error = new DatabaseError(response.statusText, response.status, {
          code: "internal",
          message: response.statusText
        });
      }
      throw error;
    }
  }
  function connect(config2) {
    return new Connection(config2);
  }
  function parseRow(fields, rawRow, cast2) {
    const row = decodeRow(rawRow);
    return fields.reduce((acc, field, ix) => {
      acc[field.name] = cast2(field, row[ix]);
      return acc;
    }, {});
  }
  function parse(result, cast2) {
    const fields = result.fields;
    const rows = result.rows ?? [];
    return rows.map((row) => parseRow(fields, row, cast2));
  }
  function decodeRow(row) {
    const values = atob(row.values);
    let offset = 0;
    return row.lengths.map((size) => {
      const width = parseInt(size, 10);
      if (width < 0)
        return null;
      const splice = values.substring(offset, offset + width);
      offset += width;
      return splice;
    });
  }
  function cast(field, value) {
    if (value === "" || value == null) {
      return value;
    }
    switch (field.type) {
      case "INT8":
      case "INT16":
      case "INT24":
      case "INT32":
      case "UINT8":
      case "UINT16":
      case "UINT24":
      case "UINT32":
      case "YEAR":
        return parseInt(value, 10);
      case "FLOAT32":
      case "FLOAT64":
        return parseFloat(value);
      case "DECIMAL":
      case "INT64":
      case "UINT64":
      case "DATE":
      case "TIME":
      case "DATETIME":
      case "TIMESTAMP":
      case "BLOB":
      case "BIT":
      case "VARBINARY":
      case "BINARY":
        return value;
      case "JSON":
        return JSON.parse(decode(value));
      default:
        return decode(value);
    }
  }

  // src/db.ts
  var workerConn = null;
  var config = {
    host: globalThis.DATABASE_HOST,
    username: globalThis.DATABASE_USER,
    password: globalThis.DATABASE_PASSWORD
  };
  function getConnection() {
    if (workerConn !== null) {
      return workerConn;
    }
    return connect(config);
  }
  function getKv() {
    return globalThis.kvStorage;
  }

  // src/cron/schedule.ts
  var fetchShopSQL = "SELECT id, url, client_id, client_secret FROM shop WHERE last_scraped_at IS NULL OR last_scraped_at < DATE_SUB(NOW(), INTERVAL 1 HOUR) ORDER BY id ASC LIMIT 1";
  async function onSchedule() {
    const con = getConnection();
    const shops = await con.execute(fetchShopSQL);
    if (shops.rows.length === 0) {
      return;
    }
    const shop = shops.rows[0];
    await con.execute("UPDATE shop SET last_scraped_at = NOW() WHERE id = ?", [shop.id]);
    const client = new import_http_client.HttpClient(new import_shop.Shop("", shop.url, "", shop.client_id, shop.client_secret));
    await client.getToken();
    const responses = await Promise.allSettled([
      client.get("/_info/config"),
      client.get("/_action/extension/installed"),
      client.post("/search/scheduled-task")
    ]);
    await con.execute("UPDATE shop SET shopware_version = ? WHERE id = ?", [
      responses[0].value.body.version,
      shop.id
    ]);
    const extensions = responses[1].value.body.map((extension) => {
      return {
        name: extension.name,
        active: extension.active,
        version: extension.version,
        latestVersion: extension.latestVersion,
        installed: extension.installedAt !== null
      };
    });
    const scheduledTasks = responses[2].value.body.data.map((task) => {
      return {
        name: task.name,
        status: task.status,
        latestVersion: task.latestVersion,
        lastExecutionTime: task.lastExecutionTime,
        nextExecutionTime: task.nextExecutionTime
      };
    });
    await con.execute("REPLACE INTO shop_scrape_info(shop_id, extensions, scheduled_task, created_at) VALUES(?, ?, ?, NOW())", [
      shop.id,
      JSON.stringify(extensions),
      JSON.stringify(scheduledTasks)
    ]);
  }

  // node_modules/itty-router/dist/itty-router.min.mjs
  function e({ base: t = "", routes: n = [] } = {}) {
    return { __proto__: new Proxy({}, { get: (e2, a, o) => (e3, ...r) => n.push([a.toUpperCase(), RegExp(`^${(t + e3).replace(/(\/?)\*/g, "($1.*)?").replace(/\/$/, "").replace(/:(\w+)(\?)?(\.)?/g, "$2(?<$1>[^/]+)$2$3").replace(/\.(?=[\w(])/, "\\.").replace(/\)\.\?\(([^\[]+)\[\^/g, "?)\\.?($1(?<=\\.)[^\\.")}/*$`), r]) && o }), routes: n, async handle(e2, ...r) {
      let a, o, t2 = new URL(e2.url);
      e2.query = Object.fromEntries(t2.searchParams);
      for (var [p, s, u] of n)
        if ((p === e2.method || "ALL" === p) && (o = t2.pathname.match(s))) {
          e2.params = o.groups;
          for (var c of u)
            if (void 0 !== (a = await c(e2.proxy || e2, ...r)))
              return a;
        }
    } };
  }

  // src/api/middleware/auth.ts
  async function validateToken(req) {
    if (req.headers.get("token") === null) {
      return new Response("Invalid token", {
        status: 401
      });
    }
    const token = req.headers.get("token");
    const result = await getKv().get(token);
    if (result === null) {
      return new Response("Invalid token", {
        status: 401
      });
    }
    const data = JSON.parse(result);
    req.userId = data.id;
  }

  // src/api/middleware/team.ts
  async function validateTeam(req) {
    const { teamId } = req.params;
    const res = await getConnection().execute("SELECT team.owner_id as ownerId FROM user_to_team INNER JOIN team ON(team.id = user_to_team.team_id) WHERE user_id = ? AND team_id = ?", [req.userId, teamId]);
    if (res.rows.length === 0) {
      return new Response("Not Found.", { status: 404 });
    }
    req.team = {
      id: teamId,
      ownerId: res.rows[0].ownerId
    };
  }
  async function validateTeamOwner(req) {
    if (req.team.ownerId !== req.userId) {
      return new Response("Forbidden.", { status: 403 });
    }
  }

  // src/api/team/create_shop.ts
  var import_http_client2 = __toESM(require_http_client());
  var import_shop2 = __toESM(require_shop());

  // src/repository/shops.ts
  var Shops = class {
    static async createShop(params) {
      const result = await getConnection().execute("INSERT INTO shop (team_id, name, url, client_id, client_secret, created_at, shopware_version) VALUES (?, ?, ?, ?, ?, NOW(), ?)", [
        params.team_id,
        params.name,
        params.shop_url,
        params.client_id,
        params.client_secret,
        params.version
      ]);
      if (result.error?.code == "ALREADY_EXISTS") {
        throw new Error("Shop already exists.");
      }
      return result.insertId;
    }
    static async deleteShop(id) {
      await getConnection().execute("DELETE FROM shop WHERE id = ?", [
        id
      ]);
      await getConnection().execute("DELETE FROM shop_scrape_info WHERE shop_id = ?", [
        id
      ]);
    }
  };

  // src/api/common/response.ts
  var NoContentResponse = class extends Response {
    constructor(headers = {}) {
      super(null, { status: 204, headers });
    }
  };
  var ErrorResponse = class extends Response {
    constructor(message, statusCode = 500) {
      super(
        JSON.stringify({ message }),
        {
          status: statusCode,
          headers: {
            "content-type": "application/json"
          }
        }
      );
    }
  };
  var JsonResponse = class extends Response {
    constructor(body, status = 200, headers = {}) {
      super(JSON.stringify(body), {
        status,
        headers: {
          "content-type": "application/json",
          ...headers
        }
      });
    }
  };

  // src/api/team/create_shop.ts
  async function createShop(req) {
    const json = await req.json();
    const { teamId } = req.params;
    if (typeof json.shop_url !== "string") {
      return new Response("Invalid shop_url.", { status: 400 });
    }
    if (typeof json.client_id !== "string") {
      return new Response("Invalid client_id.", { status: 400 });
    }
    if (typeof json.client_secret !== "string") {
      return new Response("Invalid client_secret.", { status: 400 });
    }
    const client = new import_http_client2.HttpClient(new import_shop2.Shop("", json.shop_url, "", json.client_id, json.client_secret));
    let resp;
    try {
      resp = await client.get("/_info/config");
    } catch (e2) {
      return new Response("Cannot reach shop", { status: 400 });
    }
    try {
      const id = await Shops.createShop({
        team_id: teamId,
        name: json.name || json.shop_url,
        client_id: json.client_id,
        client_secret: json.client_secret,
        shop_url: json.shop_url,
        version: resp.body.version
      });
      return new JsonResponse({ id });
    } catch (e2) {
      return new ErrorResponse(e2.message || "Unknown error");
    }
  }

  // src/api/team/delete_shop.ts
  async function deleteShop(req) {
    const { teamId, shopId } = req.params;
    const result = await getConnection().execute("SELECT shop.id, shop.url, shop.created_at, shop.shopware_version, shop.last_scraped_at, shop.shopware_version, shop_scrape_info.extensions, shop_scrape_info.scheduled_task FROM shop LEFT JOIN shop_scrape_info ON(shop_scrape_info.shop_id = shop.id) WHERE shop.id = ? AND shop.team_id = ?", [
      shopId,
      teamId
    ]);
    if (result.rows.length === 0) {
      return new Response("Not Found.", { status: 404 });
    }
    await Shops.deleteShop(shopId);
    return new NoContentResponse();
  }

  // src/api/team/get_stop.ts
  async function getShop(req) {
    const { teamId, shopId } = req.params;
    const result = await getConnection().execute("SELECT shop.id, shop.url, shop.created_at, shop.shopware_version, shop.last_scraped_at, shop.shopware_version, shop_scrape_info.extensions, shop_scrape_info.scheduled_task  FROM shop LEFT JOIN shop_scrape_info ON(shop_scrape_info.shop_id = shop.id) WHERE shop.id = ? AND shop.team_id = ?", [
      shopId,
      teamId
    ]);
    if (result.rows.length === 0) {
      return new Response("Not Found.", { status: 404 });
    }
    const shop = result.rows[0];
    shop.extensions = JSON.parse(shop.extensions);
    shop.scheduled_task = JSON.parse(shop.scheduled_task);
    return new Response(JSON.stringify(shop), { status: 200 });
  }

  // src/api/team/list_shops.ts
  async function listShops(req) {
    const { teamId } = req.params;
    const res = await getConnection().execute("SELECT id, url, created_at, last_scraped_at, shopware_version FROM shop WHERE team_id = ?", [teamId]);
    return new Response(JSON.stringify(res.rows), {
      status: 200,
      headers: {
        "Content-Type": "application/json"
      }
    });
  }

  // src/repository/teams.ts
  var Teams = class {
    static async createTeam(name, ownerId) {
      const teamInsertResult = await getConnection().execute("INSERT INTO team (name, owner_id) VALUES (?, ?)", [name, ownerId]);
      await getConnection().execute("INSERT INTO user_to_team (user_id, team_id) VALUES (?, ?)", [ownerId, teamInsertResult.insertId]);
      return teamInsertResult.insertId;
    }
    static async listMembers(teamId) {
      const result = await getConnection().execute(`SELECT
        user.id,
        user.email
    FROM user_to_team 
    INNER JOIN user ON(user.id = user_to_team.user_id)
    WHERE user_to_team.team_id = ?`, [teamId]);
      return result.rows;
    }
    static async addMember(teamId, email) {
      const member = await getConnection().execute(`SELECT id FROM user WHERE email = ?`, [email]);
      if (member.rows.length === 0) {
        return;
      }
      await getConnection().execute(`INSERT INTO user_to_team (team_id, user_id) VALUES(?, ?)`, [teamId, member.rows[0].id]);
    }
    static async removeMember(teamId, userId) {
      await getConnection().execute(`DELETE FROM user_to_team WHERE team_id = ? AND user_id = ?`, [teamId, userId]);
    }
    static async deleteTeam(teamId) {
      const shops = await getConnection().execute(`SELECT id FROM shop WHERE team_id = ?`, [teamId]);
      const shopIds = shops.rows.map((shop) => shop.id);
      await getConnection().execute(`DELETE FROM shop WHERE team_id = ?`, [teamId]);
      await getConnection().execute(`DELETE FROM shop_scrape_info WHERE shop_id IN (?)`, [shopIds]);
      await getConnection().execute(`DELETE FROM team WHERE id = ?`, [teamId]);
      await getConnection().execute(`DELETE FROM user_to_team WHERE team_id = ?`, [teamId]);
    }
  };

  // src/api/team/members.ts
  async function listMembers(req) {
    const { teamId } = req.params;
    return new JsonResponse(await Teams.listMembers(teamId));
  }
  async function addMember(req) {
    const { teamId } = req.params;
    const json = await req.json();
    if (json.email === void 0) {
      return new JsonResponse({
        message: "Missing email"
      }, 400);
    }
    await Teams.addMember(teamId, json.email);
    return new NoContentResponse();
  }
  async function removeMember(req) {
    const { teamId, userId } = req.params;
    if (userId == req.userId) {
      return new JsonResponse({
        message: "You cannot remove yourself"
      }, 400);
    }
    await Teams.removeMember(teamId, userId);
    return new NoContentResponse();
  }

  // src/api/team/team.ts
  async function deleteTeam(req) {
    await Teams.deleteTeam(req.team.id);
    return new NoContentResponse();
  }

  // src/api/team/index.ts
  var teamRouter = e({ base: "/api/team" });
  teamRouter.delete("/:teamId", validateToken, validateTeam, validateTeamOwner, deleteTeam);
  teamRouter.get("/:teamId/members", validateToken, validateTeam, listMembers);
  teamRouter.post("/:teamId/members", validateToken, validateTeam, validateTeamOwner, addMember);
  teamRouter.delete("/:teamId/members/:userId", validateToken, validateTeam, validateTeamOwner, removeMember);
  teamRouter.get("/:teamId/shops", validateToken, validateTeam, listShops);
  teamRouter.post("/:teamId/shops", validateToken, validateTeam, validateTeamOwner, createShop);
  teamRouter.get("/:teamId/shop/:shopId", validateToken, validateTeam, getShop);
  teamRouter.delete("/:teamId/shop/:shopId", validateToken, validateTeam, deleteShop);
  var team_default = teamRouter;

  // src/api/auth/confirm.ts
  async function confirmMail(req) {
    const { token } = req.params;
    const result = await getConnection().execute("SELECT id FROM user WHERE verify_code = ?", [token]);
    if (result.rows.length === 0) {
      return new ErrorResponse("Invalid confirm token", 400);
    }
    const userId = result.rows[0].id;
    await getConnection().execute("UPDATE user SET verify_code = NULL WHERE id = ?", [userId]);
    return new NoContentResponse();
  }

  // src/api/auth/login.ts
  var import_bcryptjs = __toESM(require_bcrypt());

  // src/util.ts
  function randomString(length) {
    var result = "";
    var characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    var charactersLength = characters.length;
    for (var i = 0; i < length; i++) {
      result += characters.charAt(Math.floor(Math.random() * charactersLength));
    }
    return result;
  }

  // src/api/auth/login.ts
  async function login(req) {
    const loginError = new ErrorResponse("Invalid email or password", 400);
    const json = await req.json();
    if (json.email === void 0 || json.password === void 0) {
      return new ErrorResponse("Missing email or password", 400);
    }
    const result = await getConnection().execute("SELECT id, password FROM user WHERE email = ? AND verify_code IS NULL", [json.email]);
    if (!result.rows.length) {
      return loginError;
    }
    const user = result.rows[0];
    const passwordIsValid = await import_bcryptjs.default.compare(json.password, user.password);
    if (!passwordIsValid) {
      return loginError;
    }
    const authToken = `u-${user.id}-${randomString(20)}`;
    await getKv().put(
      authToken,
      JSON.stringify({
        id: user.id
      }),
      {
        expirationTtl: 60 * 30
      }
    );
    return new JsonResponse({
      "token": authToken
    });
  }

  // src/api/auth/register.ts
  var import_bcryptjs2 = __toESM(require_bcrypt());

  // src/mail/mail.ts
  async function sendMail(mail) {
    mail.from = MAIL_FROM;
    await fetch(MAIL_URL, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "token": MAIL_SECRET
      },
      body: JSON.stringify(mail)
    });
  }
  async function sendMailConfirmToUser(email, token) {
    const mailBody = `
        <p>Hi,</p>
        <p>Thank you for registering with us. Please click the link below to confirm your email address.</p>

        <p><a href="${FRONTEND_URL}/account/confirm/${token}">Confirm email</a></p>

        <p>Best regards,</p>
        <p>FriendsOfShopware</p>
    `;
    await sendMail({
      to: email,
      subject: "Confirm your email address",
      body: mailBody
    });
  }
  async function sendMailResetPassword(email, token) {
    const mailBody = `
        <p>Hi,</p>
        <p>Please click the link below to reset your password.</p>

        <p><a href="${FRONTEND_URL}/account/reset/${token}">Reset password</a></p>

        <p>Best regards,</p>
        <p>FriendsOfShopware</p>
    `;
    await sendMail({
      to: email,
      subject: "Reset your password",
      body: mailBody
    });
  }

  // src/api/auth/register.ts
  var validateEmail = (email) => {
    return String(email).toLowerCase().match(
      /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
    );
  };
  async function register_default(req) {
    const json = await req.json();
    if (json.email === void 0 || json.password === void 0) {
      return new ErrorResponse("Missing email or password", 400);
    }
    if (!validateEmail(json.email)) {
      return new ErrorResponse("Invalid email", 400);
    }
    if (json.password.length < 8) {
      return new ErrorResponse("Password must be at least 8 characters", 400);
    }
    const result = await getConnection().execute("SELECT 1 FROM user WHERE email = ?", [json.email]);
    if (result.rows.length) {
      return new ErrorResponse("User already exists", 400);
    }
    const salt = import_bcryptjs2.default.genSaltSync(10);
    const hashedPassword = await import_bcryptjs2.default.hash(json.password, salt);
    const token = randomString(32);
    const userInsertResult = await getConnection().execute("INSERT INTO user (email, password, verify_code) VALUES (?, ?, ?)", [json.email, hashedPassword, token]);
    await Teams.createTeam(`${json.email}'s Team`, userInsertResult.insertId);
    await sendMailConfirmToUser(json.email, token);
    return new NoContentResponse();
  }

  // src/api/auth/reset.ts
  var import_bcryptjs3 = __toESM(require_bcrypt());
  async function resetPasswordMail(req) {
    const { email } = await req.json();
    const token = randomString(32);
    const result = await getConnection().execute("SELECT id FROM user WHERE email = ?", [email]);
    if (result.rows.length === 0) {
      return new NoContentResponse();
    }
    await getKv().put(`reset_${token}`, result.rows[0].id.toString());
    await sendMailResetPassword(email, token);
    return new NoContentResponse();
  }
  async function confirmResetPassword(req) {
    const { password } = await req.json();
    const { token } = req.params;
    const id = await getKv().get(`reset_${token}`);
    if (!id) {
      return new ErrorResponse("Invalid token", 400);
    }
    const salt = await import_bcryptjs3.default.genSalt(10);
    const newPassword = await import_bcryptjs3.default.hash(password, salt);
    await getConnection().execute("UPDATE user SET password = ? WHERE id = ?", [newPassword, id]);
    await getKv().delete(`reset_${token}`);
    return new NoContentResponse();
  }

  // src/api/auth/index.ts
  var authRouter = e({ base: "/api/auth" });
  authRouter.post("/register", register_default);
  authRouter.post("/login", login);
  authRouter.post("/confirm/:token", confirmMail);
  authRouter.post("/reset", resetPasswordMail);
  authRouter.post("/reset/:token", confirmResetPassword);
  var auth_default = authRouter;

  // src/repository/users.ts
  var Users = class {
    static async delete(id) {
      const ownerTeams = await getConnection().execute(`SELECT id FROM team WHERE owner_id = ?`, [id]);
      if (ownerTeams.rows.length > 0) {
        throw new Error("Cannot delete user that is the owner of a team");
      }
      await getConnection().execute(`DELETE FROM user WHERE id = ?`, [id]);
      await getConnection().execute(`DELETE FROM user_to_team WHERE user_id = ?`, [id]);
    }
  };

  // src/api/account/me.ts
  var import_bcryptjs4 = __toESM(require_bcrypt());
  var revokeTokens = async (userId) => {
    const result = await getKv().list({ prefix: `u-${userId}-` });
    for (const key of result.keys) {
      await getKv().delete(key.name);
    }
  };
  async function accountMe(req) {
    const result = await getConnection().execute("SELECT id, email, created_at FROM user WHERE id = ?", [req.userId]);
    const json = result.rows[0];
    const teamResult = await getConnection().execute("SELECT team.id, team.name, team.created_at, (team.owner_id = user_to_team.user_id) as is_owner  FROM team INNER JOIN user_to_team ON user_to_team.team_id = team.id WHERE user_to_team.user_id = ?", [req.userId]);
    json.teams = teamResult.rows;
    return new Response(JSON.stringify(json), {
      status: 200,
      headers: {
        "Content-Type": "application/json"
      }
    });
  }
  async function accountDelete(req) {
    await Users.delete(req.userId);
    await getKv().delete(req.headers.get("token"));
    return new NoContentResponse();
  }
  async function accountUpdate(req) {
    const { currentPassword, email, newPassword } = await req.json();
    const result = await getConnection().execute("SELECT id, password FROM user WHERE id = ?", [req.userId]);
    if (!result.rows.length) {
      return new ErrorResponse("User not found", 404);
    }
    const user = result.rows[0];
    if (!import_bcryptjs4.default.compareSync(currentPassword, user.password)) {
      return new ErrorResponse("Invalid password", 400);
    }
    if (email !== void 0 && !validateEmail(email)) {
      return new ErrorResponse("Invalid email", 400);
    }
    if (newPassword !== void 0 && newPassword.length < 8) {
      return new ErrorResponse("Password must be at least 8 characters", 400);
    }
    if (newPassword !== void 0) {
      const hash = import_bcryptjs4.default.hashSync(newPassword, 10);
      await revokeTokens(req.userId);
      await getConnection().execute("UPDATE user SET password = ? WHERE id = ?", [hash, req.userId]);
    }
    if (email !== void 0) {
      await getConnection().execute("UPDATE user SET email = ? WHERE id = ?", [email, req.userId]);
    }
    return new NoContentResponse();
  }

  // src/api/account/index.ts
  var accountRouter = e({ base: "/api/account" });
  accountRouter.get("/me", validateToken, accountMe);
  accountRouter.patch("/me", validateToken, accountUpdate);
  accountRouter.delete("/me", validateToken, accountDelete);
  var account_default = accountRouter;

  // src/router.ts
  var router = e();
  router.all("/api/account/*", account_default.handle);
  router.all("/api/team/*", team_default.handle);
  router.all("/api/auth/*", auth_default.handle);
  router.all("*", () => new JsonResponse({ message: "Not found" }, 404));
  var router_default = router;

  // src/index.ts
  addEventListener("fetch", (event) => {
    event.respondWith(router_default.handle(event.request));
  });
  addEventListener("scheduled", (event) => {
    event.waitUntil(onSchedule());
  });
})();
/**
 * @license bcrypt.js (c) 2013 Daniel Wirtz <dcode@dcode.io>
 * Released under the Apache License, Version 2.0
 * see: https://github.com/dcodeIO/bcrypt.js for details
 */
