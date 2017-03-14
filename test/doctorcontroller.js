angular.module('editorPanelApp').controller('doctorscontroller',['$scope', '$rootScope', 'pagedb', 'util_service', 'doctor_service', 'specialty_service','template_service','uploadify_service', 'doctorFactory', 'location_service', function ($scope, $rootScope, pagedb, util_service, doctor_service, specialty_service,
	template_service, uploadify_service, doctorFactory, location_service) {
	"use strict"
	$scope.uploads = {};
	$scope.doctors = [];
	$scope.sDoctor = new doctorFactory();

	$scope.base_url = MN.base_url;
	$scope.status = '';
	$scope.notvalid = false;
	$scope.deleteIsConfirmed = false;
	$scope.askForDeleteConfirm = false;
	$scope.redactorObject = {
		buttons: [ /*'formatting', '|',*/ 'fontcolor', /*'backcolor',*/ 'bold', 'italic', 'underline', '|',
			'unorderedlist', 'orderedlist', 'outdent', 'indent', '|', /*'image', 'video',*/ 'table', 'link', '|',
			'alignment', 'horizontalrule', '|', 'undo', 'redo'
		],
		focus: false,
		fixed: false,
		fixedBox: true,
		observeImages: false,
		convertDivs: false
	};
	$scope.biographyElement = $("#doctor_biography");
	$scope.medschoolElement = $("#education");
	$scope.boardCertificatesElement = $("#board_certificates");
	$scope.residencyElement = $("#residency");
	$scope.organizationsElement = $("#organizations");

	$scope.findDoctorLocation = function () {
		if (!$scope.sDoctor || !$scope.locations) {
			return;
		}
		for (var i = $scope.locations.length - 1; i >= 0; i--) {
			if ($scope.sDoctor.location_id == $scope.locations[i].id) {
				$scope.sLocation = $scope.locations[i];
			}
		}
	};

	$scope.getDoctorSpecialties = function () {
		doctor_service.getDoctorSpecialties($scope.sDoctor.id);
	}

	$scope.init = function () {

		util_service.preventRedactorIEIssue();

		var eventlisters = true;

		$scope.uploadifyFallBack = uploadifyFallBack;
		uploadify_service.init($scope, '#photo-uploadify', 'upload-photo-button.png', 138, 37, 'doctor');
		//uploadify_service.init($scope, '#cv-uploadify', 'upload-cv-button.png', 109, 37);

		$scope.long_info_text =
			"Enter information about your practice’s providers. Use this panel to add a photo, biography and more information for each provider.  Also use this panel to remove a provider.";
		$scope.short_info_text = util_service.shortenInfoText($scope.long_info_text);

		// Place the short info text in the box
		$scope.show_long_info = false;

		$scope.$on('doctor/getdoctorspecialties', function (e, data) {
			$scope.sDoctor.specialties = data;
		});

		$scope.$on('doctor/getdoctor', function (e, data) {
			$scope.doctors = [];

			if (angular.isArray(data)) {
				for (var i = 0; i < data.length; i++) {
					var factoryObj = new doctorFactory();
					$scope.doctors.push(factoryObj.extend(data[i]));
				}
			} else {
				// no doctors perhaps
			}
			doctor_service.setRamDoctors($scope.doctors);
		});
		//Doctor has been saved to DB now update their row in DB and update current doctor object
		$scope.$on('doctor/uploadSavedToDoctor', function (e, data) {
			$scope.sDoctor.setUploadData(data);
		});
		//Doctor has not been saved yet
		$scope.$on('doctor/uploadSaved', function (e, data) {
			$scope.sDoctor.setUploadData(data);
		});

		$scope.$on('doctor/saveDoctorPhotoAndCV', function (e, data) {

			if ($scope.sDoctor.hasOwnProperty('id')) {
				// Save upload information to newly created doctor
				$scope.savePhotoOrCVByDoctor(data);
				var doc_pg_data = {
					doc_id: $scope.sDoctor.id,
					useCv: $scope.sDoctor.useCv
				};
				pagedb.buildPageByType('doctor', doc_pg_data).then(function (data) {
					$rootScope.$emit('editor/refreshDetailedPricing');
					util_service.refresh(data.page_id);
				});
			} else {
				// Save upload data only but don't associate to a doctor
				doctor_service.saveUploadData();
			}
		});

		$scope.$on('doctor/savedoctor', function (e, data) {

			var failed = false;
			var endedOn = 0;
			for (var i = 0; i < data.length; i++) {
				if (data[i].status == 0) {
					failed = true;
					endedOn = i;
					break;
				}
			}

			if (!failed) {
				$scope.hasError = false;
				$scope.status = data[endedOn].message;
			} else {
				$scope.status = false;
				$scope.hasError = true;
				$scope.errorMsg = data[endedOn].message;
			}

			doctor_service.getdoctors();

			$('.save-button').html("Save").removeAttr("disabled");

			if ($scope.sDoctor.id == undefined) {
				$scope.currentDoctorId = data[data.length - 1].data.doctor_id;
			} else {
				$scope.currentDoctorId = $scope.sDoctor.id;
			}

			if (!failed) {
				//Set current doctors id
				var lastDoctorIndex = data.length - 1;
				$scope.sDoctor.id = (jQuery.isEmptyObject($scope.sDoctor)) ? null : data[lastDoctorIndex].data.doctor_id;

				var doctors_array = [];

				for (var j = data.length - 1; j >= 0; j--) {
					doctors_array.push({
						doc_id: data[j].data.doctor_id,
						useCv: false
					});
				}
				pagedb.buildMultiDoctorPage(doctors_array);
				$scope.sDoctor = new doctorFactory();
			}


		});

		$scope.$on('doctor/doctorpagesbuilt', function (e, data) {
			for (var i = data.length - 1; i >= 0; i--) {
				if (data[i].doc_id == $scope.currentDoctorId) {
					util_service.refresh(data[i].page_id);
					break;
				}
			}
			pagedb.buildDoctorLandingPage($scope.doctors.length).then(function (data) {
				$rootScope.$emit('editor/refreshDetailedPricing');
			});
		});

		$scope.$on('doctor/removedoctor', function (e, data) {
			$rootScope.$emit('editor/refreshDetailedPricing');
			doctor_service.getdoctors();
			if (data.status === 1) {
				$scope.hasError = false;
				$scope.status = data.message;
			} else if (data.status === 0) {
				$scope.status = false;
				$scope.hasError = true;
				$scope.errorMsg = data.message;
			}
		});

		$scope.$on('location/getlocations', function (e, data) {

			if (angular.isArray(data) && data.length > 0) {
				$scope.locations = location_service.getLocations();
				$scope.findDoctorLocation();
			} else {
				$scope.locations = [];
			}
		});

		// get specialties
		specialty_service.getSpecialtiesUnformatted().then(

			function (data) {
				$scope.specialties = template_service.formatSpecialties(data, true);
			},
			function (err) {
				alert(err);
			});
		doctor_service.getdoctors();
		location_service.getlocations();

		// Set placeholders for inputs
		$('input, textarea').placeholder();


		/**
		 * Set a variable that says if the redactor button has been
		 * pressed to prevent the alert popup for leaving the page.
		 * Leave it set for half of a second.
		 */
		$(document).ready(function () {
			$("form").on("click", ".redactor_toolbar li a", function () {
				util_service.setRedactorButtonPressed(true);
				setTimeout(function () {
					util_service.setRedactorButtonPressed(false);
				}, 500);
			});
		});

		util_service.initCompleteTabsSaveButtonTooltip();

	}();

	/**
	 * Update the current doctor in the doctor service list
	 */
	$scope.updateDoctor = function () {
		var doctors = [];
		doctor_service.addUpdatedDoctors($scope.sDoctor);
		/**
		 * Save Current Form
		 */
		if (doctor_service.isDoctorUnique($scope.sDoctor)) {
			doctor_service.addDoctor($scope.sDoctor);
		} else {
			// update object in doctors array
			doctors = doctor_service.getRamDoctors();
			for (var i = 0; i < doctors.length; i++) {
				if ($scope.sDoctor.guid == doctors[i]['guid'] || $scope.sDoctor.id == doctors[i]['id']) {
					doctors.splice(i, 1, $scope.sDoctor);
				}
			}
			doctor_service.setRamDoctors(doctors);
			$scope.doctors = doctors;
		}
	};


	$scope.saveDoctors = function (event) {
		

		if($scope.doctors.length >= 1 && MN.build_type == 'provider'){
			if($scope.doctors.length == 1 && $scope.doctors[0].site_id != null){
				
			}else{
				$scope.$emit('upgradeToPracticePopover',{ev:event});
				return;
			}
		}
		var validation = $scope.validate();
		if (validation === true) {
			return;
		}
		if (validation === 'pristine') {
			$scope.status = false;
			$scope.hasError = true;
			$scope.errorMsg = "Please enter Doctor Profile information";
			return;
		}

		// Get the values of the redactor text areas
		$scope.sDoctor.biography = $scope.biographyElement.getCode();
		$scope.sDoctor.medschool = $scope.medschoolElement.getCode();
		$scope.sDoctor.board_certificates = $scope.boardCertificatesElement.getCode();
		$scope.sDoctor.residency = $scope.residencyElement.getCode();
		$scope.sDoctor.organizations = $scope.organizationsElement.getCode();

		$('.save-button').html("Saving...").attr("disabled", "disabled");

		$scope.updateDoctor();
		doctor_service.save();
	};


	$scope.validate = function () {

		var keys = {
			'doctor_firstname': 'First Name',
			'doctor_lastname': 'Last Name',
			'doctor_title': 'Title',
			'doctor_gender': 'Gender',
			'doctor_specialty': 'Specialty',
			'doctor_sub_specialty': 'Secondary Specialty',
			'doctor_biography': 'Biography'
		};

		return util_service.formValidator($scope, keys, 'form');
	};



	$scope.addDoctor = function () {
		// Reset messages
		$scope.status = false;
		$scope.hasError = false;

		var validation = $scope.validate();
		if (validation === true) {
			return;
		}
		if (validation === 'pristine') {
			$scope.hasError = true;
			$scope.errorMsg = "Please enter Doctor Profile information";
			return;
		}
		$scope.updateDoctor();
		$scope.sDoctor = new doctorFactory();
	};

	$scope.resetDoctorForm = function () {
		$scope.status = false;
		$scope.hasError = false;
		$scope.updateDoctor();
		$scope.sDoctor = new doctorFactory();
	};

	$scope.removedoctor = function () {
		if (!$scope.sDoctor || !$scope.sDoctor.id) {
			$scope.status = false;
			$scope.hasError = true;
			$scope.errorMsg = "Please select a Doctor Profile to delete";
			return;
		}

		if ($scope.deleteIsConfirmed === true) {
			var doc_id = $scope.sDoctor.id;
			doctor_service.removedoctor($scope.sDoctor);
			$scope.resetDoctorForm();
			$scope.deleteIsConfirmed = false;
			$scope.askForDeleteConfirm = false;
			pagedb.deleteDoctorPage(doc_id).then(function (data) {
				pagedb.buildDoctorLandingPage($scope.doctors.length).then(function (data) {
					util_service.refresh(data.page_id);
				});
			});
		} else if ($scope.deleteIsConfirmed === false) {
			$scope.askForDeleteConfirm = true;
		}
	};

	$scope.confirmDelete = function () {
		$scope.deleteIsConfirmed = true;
	};

	$scope.cancelDelete = function () {
		$scope.deleteIsConfirmed = false;
		$scope.askForDeleteConfirm = false;
	};

	    
	$scope.savePhotoOrCVByDoctor = function (data) {
		if ($scope.sDoctor.hasOwnProperty('id')) {            
			var post = 'site_id=' + MN.site_id + "&";            
			post += 'entity_id=' + $scope.sDoctor.id + "&";            
			post += 'type=' + "doctor" + "&";            
			post += 'filename=' + data[0].name;       
			doctor_service.uploadPhotoOrCV(post);        
		}    
	};

	$scope.deleteMedia = function (media_id, column_type) {

		$('.save-button').html("Saving...").attr("disabled", "disabled");
		if ($scope.sDoctor.hasOwnProperty('id')) {
			var entity_id = $scope.sDoctor.id;
		} else {
			var entity_id = false;
		} 
		var post = 'media_id=' + media_id + "&";
		post += 'entity_id=' + entity_id + "&";
		post += 'type=' + "doctor" + "&";
		post += 'column_type=' + column_type + "&";

		doctor_service.deleteMedia(post);
		if (column_type == "photo_media_id") {
			$scope.sDoctor.resetImageUpload();

			$('.save-button').html("Save").removeAttr("disabled");
			$scope.hasError = false;
			$scope.status = "Image deletion saved";
		} else {
			$scope.sDoctor.resetCVupload();
		}
	};
}]);