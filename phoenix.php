<?php
namespace Phoenix;
/**
 * author 逆雪寒
 * version 0.9.1
 */
class Phoenix {

	private static $instance = NULL;
	private $host = '';
	private $timeOutSecond = 2;


	private function __construct($host,$port){
		$this->host($host,$port);
	}
	 
	public function __clone(){
		trigger_error('Clone is not allow!',E_USER_ERROR);
	}

	private function host($host,$port) {
		$this->host = "http://" . $host . ":" . $port."/";
	}

	public static function  factory($host = 'localhost',$port = 8888) {
		if(is_null(self::$instance)) {
			self::$instance = new self($host,$port);
		}
		return self::$instance;
	}

	protected function request($do,$parameter = '',$method = 'GET') {
		$query = '';
		$ch = curl_init(); 
		$query = http_build_query($parameter);

		if($method != 'GET') {
			curl_setopt($ch, CURLOPT_POSTFIELDS,$query);
			curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $method);
			curl_setopt($ch, CURLOPT_URL, $this->host.$do);
		}else{
			$query = "?" . $query;
			curl_setopt($ch, CURLOPT_URL, $this->host.$do.$query);
		}
		curl_setopt($ch, CURLOPT_TIMEOUT,$this->timeOutSecond);
		curl_setopt($ch, CURLOPT_RETURNTRANSFER,true);

		$data = curl_exec($ch);
		if(curl_errno($ch)){ 
			return false;
		}
		curl_close($ch);
		return $data;
	}

	private $groupName = '';

	/**
	 * 设置超时
	 */
	public function setTimeOut($second){
		$this->timeOutSecond = $second;
		return $this;
	}

	/**
	 * 设置组
	 *
	 * @param string $queueName 队列名
	 */
	public function setGroupName($groupName){
		$this->groupName = $groupName;
		return $this;
	}

	/**
	 * 打点
	 */
	public function dota($title,$timeUsed) {
		if($this->groupName == ""){
			return false;
		}

		$result = $this->request('dota',[
			'groupName' => $this->groupName,
			'timeUsed' 		=> $timeUsed,
			'title'		=> $title
		],'POST');

		if($result == "OK") {
			return true;
		}
		return false;
	}
}

$phoenix = Phoenix::factory();
var_dump($phoenix->setTimeOut(3)->setGroupName('bee')->dota("/v1.1.1/timeline",0.22323434));

