// Libraries
import React, {PureComponent} from 'react'
// Components
import {ErrorHandling} from 'src/shared/decorators/errors'

// Types
import {StepStatus} from 'src/clockface/constants/wizard'
import {Button, ComponentColor, ComponentSize} from 'src/clockface'
import {OnboardingStepProps} from 'src/onboarding/containers/OnboardingWizard'

@ErrorHandling
class InitStep extends PureComponent<OnboardingStepProps> {
  public render() {
    return (
      <div className="onboarding-step">
        <div className="splash-logo primary" />
        <h3 className="wizard-step-title">Welcome to InfluxData </h3>
        <p>"Start using the InfluxData platform in a few easy steps"</p>
        <div className="wizard-button-bar">
          <Button
            color={ComponentColor.Primary}
            text="Get Started"
            size={ComponentSize.Large}
            onClick={this.handleNext}
          />
        </div>
      </div>
    )
  }

  private handleNext = () => {
    const {handleSetStepStatus, currentStepIndex} = this.props
    handleSetStepStatus(currentStepIndex, StepStatus.Complete)
    this.handleIncrement()
  }

  private handleIncrement = () => {
    const {handleSetCurrentStep, currentStepIndex} = this.props
    handleSetCurrentStep(currentStepIndex + 1)
  }
}

export default InitStep
