package constant

const RedisGetKeyErr string = "从redis中获取key出现错误"
const RedisSetKeyErr string = "向redis存放key时出现错误"
const JsonMarshalErr string = "转成字节时出现错误"
const JsonUnMarshalErr string = "字节转成对象时出现错误"
const RabbitmqSendErr string = "rabbitmq发送失败"
const HashEncodeErr string = "使用hash算法加密错误"
const MysqlUpdateErr string = "mysql更新实体失败"
const MysqlUCreateErr string = "mysql创建实体失败"
const MysqlUDeleteErr string = "mysql删除实体失败"
const MysqlUFindErr string = "mysql查找实体失败"
const JWTCreateErr string = "创建JWT令牌失败"
const JWTGetUserIDErr string = "从jwt中获取id出现错误"
const UserIdConvertIntErr = "从jwt中获取id转换成int时错误"
const ImageUpLoadErr = "获取图片信息错误"
const GenerateUUIDErr = "生成uuid出现错误"
const FileOpenErr = "文件打开错误"
const FileCloseErr = "文件关闭错误"
const MinioUploadErr = "上传minio文件时出现错误"
const MinioGetErr = "获取minio文件时出现错误"
const MinioDeleteErr = "删除minio文件时出现错误"
const ObjectCopyToWriterErr = "获取minio对象的写入ctx的writer中出现错误"
